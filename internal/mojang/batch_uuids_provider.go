package mojang

import (
	"strings"
	"sync"
	"time"

	"github.com/elyby/chrly/internal/utils"
)

type BatchUuidsProvider struct {
	UsernamesToUuidsEndpoint func(usernames []string) ([]*ProfileInfo, error)
	batch                    int
	delay                    time.Duration
	fireOnFull               bool

	queue       *utils.Queue[*job]
	fireChan    chan any
	stopChan    chan any
	onFirstCall sync.Once
}

func NewBatchUuidsProvider(
	endpoint func(usernames []string) ([]*ProfileInfo, error),
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
	ResultChan chan<- *jobResult
}

type jobResult struct {
	Profile *ProfileInfo
	Error   error
}

func (ctx *BatchUuidsProvider) GetUuid(username string) (*ProfileInfo, error) {
	resultChan := make(chan *jobResult)
	n := ctx.queue.Enqueue(&job{username, resultChan})
	if ctx.fireOnFull && n%ctx.batch == 0 {
		ctx.fireChan <- struct{}{}
	}

	ctx.onFirstCall.Do(ctx.startQueue)

	result := <-resultChan

	return result.Profile, result.Error
}

func (ctx *BatchUuidsProvider) StopQueue() {
	close(ctx.stopChan)
}

func (ctx *BatchUuidsProvider) startQueue() {
	go func() {
		for {
			t := time.NewTimer(ctx.delay)
			select {
			case <-ctx.stopChan:
				return
			case <-t.C:
				go ctx.fireRequest()
			case <-ctx.fireChan:
				t.Stop()
				go ctx.fireRequest()
			}
		}
	}()
}

func (ctx *BatchUuidsProvider) fireRequest() {
	jobs, _ := ctx.queue.Dequeue(ctx.batch)
	if len(jobs) == 0 {
		return
	}

	usernames := make([]string, len(jobs))
	for i, job := range jobs {
		usernames[i] = job.Username
	}

	profiles, err := ctx.UsernamesToUuidsEndpoint(usernames)
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
