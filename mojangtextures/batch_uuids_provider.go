package mojangtextures

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/elyby/chrly/api/mojang"
)

type jobResult struct {
	Profile *mojang.ProfileInfo
	Error   error
}

type job struct {
	Username    string
	RespondChan chan *jobResult
}

type jobsQueue struct {
	lock  sync.Mutex
	items []*job
}

func newJobsQueue() *jobsQueue {
	return &jobsQueue{
		items: []*job{},
	}
}

func (s *jobsQueue) Enqueue(job *job) int {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items = append(s.items, job)

	return len(s.items)
}

func (s *jobsQueue) Dequeue(n int) ([]*job, int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.items)
	if n > l {
		n = l
	}

	items := s.items[0:n]
	s.items = s.items[n:l]

	return items, l - n
}

var usernamesToUuids = mojang.UsernamesToUuids

type JobsIteration struct {
	Jobs  []*job
	Queue int
	c     chan struct{}
}

func (j *JobsIteration) Done() {
	if j.c != nil {
		close(j.c)
	}
}

type BatchUuidsProviderStrategy interface {
	Queue(job *job)
	GetJobs(abort context.Context) <-chan *JobsIteration
}

type PeriodicStrategy struct {
	Delay time.Duration
	Batch int
	queue *jobsQueue
	done  chan struct{}
}

func NewPeriodicStrategy(delay time.Duration, batch int) *PeriodicStrategy {
	return &PeriodicStrategy{
		Delay: delay,
		Batch: batch,
		queue: newJobsQueue(),
	}
}

func (ctx *PeriodicStrategy) Queue(job *job) {
	ctx.queue.Enqueue(job)
}

func (ctx *PeriodicStrategy) GetJobs(abort context.Context) <-chan *JobsIteration {
	ch := make(chan *JobsIteration)
	go func() {
		for {
			select {
			case <-abort.Done():
				close(ch)
				return
			case <-time.After(ctx.Delay):
				jobs, queueLen := ctx.queue.Dequeue(ctx.Batch)
				jobDoneChan := make(chan struct{})
				ch <- &JobsIteration{jobs, queueLen, jobDoneChan}
				<-jobDoneChan
			}
		}
	}()

	return ch
}

type FullBusStrategy struct {
	Delay     time.Duration
	Batch     int
	queue     *jobsQueue
	busIsFull chan bool
}

func NewFullBusStrategy(delay time.Duration, batch int) *FullBusStrategy {
	return &FullBusStrategy{
		Delay:     delay,
		Batch:     batch,
		queue:     newJobsQueue(),
		busIsFull: make(chan bool),
	}
}

func (ctx *FullBusStrategy) Queue(job *job) {
	n := ctx.queue.Enqueue(job)
	if n % ctx.Batch == 0 {
		ctx.busIsFull <- true
	}
}

// Формально, это описание логики водителя маршрутки xD
func (ctx *FullBusStrategy) GetJobs(abort context.Context) <-chan *JobsIteration {
	ch := make(chan *JobsIteration)
	go func() {
		for {
			t := time.NewTimer(ctx.Delay)
			select {
			case <-abort.Done():
				close(ch)
				return
			case <-t.C:
				ctx.sendJobs(ch)
			case <-ctx.busIsFull:
				t.Stop()
				ctx.sendJobs(ch)
			}
		}
	}()

	return ch
}

func (ctx *FullBusStrategy) sendJobs(ch chan *JobsIteration) {
	jobs, queueLen := ctx.queue.Dequeue(ctx.Batch)
	ch <- &JobsIteration{jobs, queueLen, nil}
}

type BatchUuidsProvider struct {
	context     context.Context
	emitter     Emitter
	strategy    BatchUuidsProviderStrategy
	onFirstCall sync.Once
}

func NewBatchUuidsProvider(
	context context.Context,
	strategy BatchUuidsProviderStrategy,
	emitter Emitter,
) *BatchUuidsProvider {
	return &BatchUuidsProvider{
		context:  context,
		emitter:  emitter,
		strategy: strategy,
	}
}

func (ctx *BatchUuidsProvider) GetUuid(username string) (*mojang.ProfileInfo, error) {
	ctx.onFirstCall.Do(ctx.startQueue)

	resultChan := make(chan *jobResult)
	ctx.strategy.Queue(&job{username, resultChan})
	ctx.emitter.Emit("mojang_textures:batch_uuids_provider:queued", username)

	result := <-resultChan

	return result.Profile, result.Error
}

func (ctx *BatchUuidsProvider) startQueue() {
	go func() {
		jobsChan := ctx.strategy.GetJobs(ctx.context)
		for {
			select {
			case <-ctx.context.Done():
				return
			case iteration := <-jobsChan:
				go func() {
					ctx.performRequest(iteration)
					iteration.Done()
				}()
			}
		}
	}()
}

func (ctx *BatchUuidsProvider) performRequest(iteration *JobsIteration) {
	usernames := make([]string, len(iteration.Jobs))
	for i, job := range iteration.Jobs {
		usernames[i] = job.Username
	}

	ctx.emitter.Emit("mojang_textures:batch_uuids_provider:round", usernames, iteration.Queue)
	if len(usernames) == 0 {
		return
	}

	profiles, err := usernamesToUuids(usernames)
	ctx.emitter.Emit("mojang_textures:batch_uuids_provider:result", usernames, profiles, err)
	for _, job := range iteration.Jobs {
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

		job.RespondChan <- response
		close(job.RespondChan)
	}
}
