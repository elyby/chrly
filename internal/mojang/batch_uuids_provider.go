package mojang

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/SentimensRG/ctx/mergectx"

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
}

func NewBatchUuidsProvider(
	endpoint UsernamesToUuidsEndpoint,
	batchSize int,
	awaitDelay time.Duration,
	fireOnFull bool,
) *BatchUuidsProvider {
	return &BatchUuidsProvider{
		UsernamesToUuidsEndpoint: endpoint,
		stopChan:                 make(chan any),
		batch:                    batchSize,
		delay:                    awaitDelay,
		fireOnFull:               fireOnFull,
		queue:                    utils.NewQueue[*job](),
		fireChan:                 make(chan any),
	}
}

type job struct {
	Username   string
	Ctx        context.Context
	ResultChan chan<- *jobResult
}

type jobResult struct {
	Profile *ProfileInfo
	Error   error
}

func (p *BatchUuidsProvider) GetUuid(ctx context.Context, username string) (*ProfileInfo, error) {
	resultChan := make(chan *jobResult)
	n := p.queue.Enqueue(&job{username, ctx, resultChan})
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
	jobs := make([]*job, 0, p.batch)
	n := p.batch
	for {
		foundJobs, left := p.queue.Dequeue(n)
		for i := range foundJobs {
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

	ctx := context.Background()
	usernames := make([]string, len(jobs))
	for i, job := range jobs {
		usernames[i] = job.Username
		ctx = mergectx.Join(ctx, job.Ctx)
	}

	profiles, err := p.UsernamesToUuidsEndpoint(ctx, usernames)
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
