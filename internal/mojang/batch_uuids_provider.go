package mojang

import (
	"context"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"

	"ely.by/chrly/internal/otel"
	"ely.by/chrly/internal/utils"
)

type UsernamesToUuidsEndpoint func(ctx context.Context, usernames []string) ([]*ProfileInfo, error)

type BatchUuidsProvider struct {
	UsernamesToUuidsEndpoint
	batch      int
	delay      time.Duration
	fireOnFull bool

	queue       *utils.Queue[*job]
	fireChan    chan any
	stopChan    chan any
	onFirstCall sync.Once
	metrics     *batchUuidsProviderMetrics
}

func NewBatchUuidsProvider(
	endpoint UsernamesToUuidsEndpoint,
	batchSize int,
	awaitDelay time.Duration,
	fireOnFull bool,
) (*BatchUuidsProvider, error) {
	queue := utils.NewQueue[*job]()

	metrics, err := newBatchUuidsProviderMetrics(otel.GetMeter(), queue)
	if err != nil {
		return nil, err
	}

	return &BatchUuidsProvider{
		UsernamesToUuidsEndpoint: endpoint,
		stopChan:                 make(chan any),
		batch:                    batchSize,
		delay:                    awaitDelay,
		fireOnFull:               fireOnFull,
		queue:                    queue,
		fireChan:                 make(chan any),
		metrics:                  metrics,
	}, nil
}

type job struct {
	Username    string
	Ctx         context.Context
	QueuingTime time.Time
	ResultChan  chan<- *jobResult
}

type jobResult struct {
	Profile *ProfileInfo
	Error   error
}

func (p *BatchUuidsProvider) GetUuid(ctx context.Context, username string) (*ProfileInfo, error) {
	resultChan := make(chan *jobResult)
	n := p.queue.Enqueue(&job{username, ctx, time.Now(), resultChan})
	if p.fireOnFull && n%p.batch == 0 {
		p.fireChan <- struct{}{}
	}

	p.onFirstCall.Do(p.startQueue)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultChan:
		return result.Profile, result.Error
	}
}

func (p *BatchUuidsProvider) StopQueue() {
	close(p.stopChan)
}

func (p *BatchUuidsProvider) startQueue() {
	go func() {
		for {
			t := time.NewTimer(p.delay)
			select {
			case <-p.stopChan:
				return
			case <-t.C:
				go p.fireRequest()
			case <-p.fireChan:
				t.Stop()
				go p.fireRequest()
			}
		}
	}()
}

func (p *BatchUuidsProvider) fireRequest() {
	// Since this method is an aggregator, it uses its own context to manage its lifetime
	reqCtx := context.Background()
	jobs := make([]*job, 0, p.batch)
	n := p.batch
	for {
		foundJobs, left := p.queue.Dequeue(n)
		for i := range foundJobs {
			p.metrics.QueueTime.Record(reqCtx, float64(time.Since(foundJobs[i].QueuingTime).Milliseconds()))
			if foundJobs[i].Ctx.Err() != nil {
				// If the job context has already ended, its result will be returned in the GetUuid method
				close(foundJobs[i].ResultChan)

				foundJobs[i] = foundJobs[len(foundJobs)-1]
				foundJobs = foundJobs[:len(foundJobs)-1]
			}
		}

		jobs = append(jobs, foundJobs...)
		if len(jobs) != p.batch && left != 0 {
			n = p.batch - len(jobs)
			continue
		}

		break
	}

	if len(jobs) == 0 {
		return
	}

	usernames := make([]string, len(jobs))
	for i, job := range jobs {
		usernames[i] = job.Username
	}

	p.metrics.Requests.Add(reqCtx, 1)
	p.metrics.BatchSize.Record(reqCtx, int64(len(usernames)))

	profiles, err := p.UsernamesToUuidsEndpoint(reqCtx, usernames)
	for _, job := range jobs {
		response := &jobResult{}
		if err == nil {
			// The profiles in the response aren't ordered, so we must search each username over full array
			for _, profile := range profiles {
				if strings.EqualFold(job.Username, profile.Name) {
					response.Profile = profile
					break
				}
			}
		} else {
			response.Error = err
		}

		job.ResultChan <- response
		close(job.ResultChan)
	}
}

func newBatchUuidsProviderMetrics(meter metric.Meter, queue *utils.Queue[*job]) (*batchUuidsProviderMetrics, error) {
	m := &batchUuidsProviderMetrics{}
	var errors, err error

	m.Requests, err = meter.Int64Counter(
		"chrly.mojang.uuids.batch.request.sent",
		metric.WithDescription("Number of UUIDs requests sent to Mojang API"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	m.BatchSize, err = meter.Int64Histogram(
		"chrly.mojang.uuids.batch.request.batch_size",
		metric.WithDescription("The number of usernames in the query"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	m.QueueLength, err = meter.Int64ObservableGauge(
		"chrly.mojang.uuids.batch.queue.length",
		metric.WithDescription("Number of tasks in the queue waiting for execution"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(queue.Len()))
			return nil
		}),
	)
	errors = multierr.Append(errors, err)

	m.QueueTime, err = meter.Float64Histogram(
		"chrly.mojang.uuids.batch.queue.lag",
		metric.WithDescription("Lag between placing a job in the queue and starting its processing"),
		metric.WithUnit("ms"),
	)
	errors = multierr.Append(errors, err)

	return m, errors
}

type batchUuidsProviderMetrics struct {
	Requests    metric.Int64Counter
	BatchSize   metric.Int64Histogram
	QueueLength metric.Int64ObservableGauge
	QueueTime   metric.Float64Histogram
}
