package mojangtextures

import (
	"strings"
	"sync"
	"time"

	"github.com/mono83/slf/wd"

	"github.com/elyby/chrly/api/mojang"
)

type jobResult struct {
	profile *mojang.ProfileInfo
	error   error
}

type jobItem struct {
	username    string
	respondChan chan *jobResult
}

type jobsQueue struct {
	lock  sync.Mutex
	items []*jobItem
}

func (s *jobsQueue) New() *jobsQueue {
	s.items = []*jobItem{}
	return s
}

func (s *jobsQueue) Enqueue(t *jobItem) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items = append(s.items, t)
}

func (s *jobsQueue) Dequeue(n int) []*jobItem {
	s.lock.Lock()
	defer s.lock.Unlock()

	if n > s.size() {
		n = s.size()
	}

	items := s.items[0:n]
	s.items = s.items[n:len(s.items)]

	return items
}

func (s *jobsQueue) Size() int {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.size()
}

func (s *jobsQueue) size() int {
	return len(s.items)
}

var usernamesToUuids = mojang.UsernamesToUuids
var forever = func() bool {
	return true
}

type BatchUuidsProvider struct {
	IterationDelay time.Duration
	IterationSize  int
	Logger         wd.Watchdog

	onFirstCall sync.Once
	queue       jobsQueue
}

func (ctx *BatchUuidsProvider) GetUuid(username string) (*mojang.ProfileInfo, error) {
	ctx.onFirstCall.Do(func() {
		ctx.queue.New()
		ctx.startQueue()
	})

	resultChan := make(chan *jobResult)
	ctx.queue.Enqueue(&jobItem{username, resultChan})
	ctx.Logger.IncCounter("mojang_textures.usernames.queued", 1)

	result := <-resultChan

	return result.profile, result.error
}

func (ctx *BatchUuidsProvider) startQueue() {
	go func() {
		time.Sleep(ctx.IterationDelay)
		for forever() {
			start := time.Now()
			ctx.queueRound()
			elapsed := time.Since(start)
			ctx.Logger.RecordTimer("mojang_textures.usernames.round_time", elapsed)
			time.Sleep(ctx.IterationDelay)
		}
	}()
}

func (ctx *BatchUuidsProvider) queueRound() {
	queueSize := ctx.queue.Size()
	jobs := ctx.queue.Dequeue(ctx.IterationSize)
	ctx.Logger.UpdateGauge("mojang_textures.usernames.queue_size", int64(queueSize-len(jobs)))
	ctx.Logger.UpdateGauge("mojang_textures.usernames.iteration_size", int64(len(jobs)))
	if len(jobs) == 0 {
		return
	}

	var usernames []string
	for _, job := range jobs {
		usernames = append(usernames, job.username)
	}

	profiles, err := usernamesToUuids(usernames)
	for _, job := range jobs {
		go func(job *jobItem) {
			response := &jobResult{}
			if err != nil {
				response.error = err
			} else {
				// The profiles in the response aren't ordered, so we must search each username over full array
				for _, profile := range profiles {
					if strings.EqualFold(job.username, profile.Name) {
						response.profile = profile
						break
					}
				}
			}

			job.respondChan <- response
		}(job)
	}
}
