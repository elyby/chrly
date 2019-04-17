package queue

import (
	"strings"
	"sync"
	"time"

	"github.com/elyby/chrly/api/mojang"
)

var usernamesToUuids = mojang.UsernamesToUuids
var uuidToTextures = mojang.UuidToTextures
var delay = time.Second

type JobsQueue struct {
	Storage Storage

	onFirstCall sync.Once
	queue       jobsQueue
}

func (ctx *JobsQueue) GetTexturesForUsername(username string) *mojang.SignedTexturesResponse {
	ctx.onFirstCall.Do(func() {
		ctx.queue.New()
		ctx.startQueue()
	})

	resultChan := make(chan *mojang.SignedTexturesResponse)
	// TODO: prevent of adding the same username more than once
	ctx.queue.Enqueue(&jobItem{username, resultChan})

	return <-resultChan
}

func (ctx *JobsQueue) startQueue() {
	go func() {
		time.Sleep(delay)
		for true {
			start := time.Now()
			ctx.queueRound()
			time.Sleep(delay - time.Since(start))
		}
	}()
}

func (ctx *JobsQueue) queueRound() {
	if ctx.queue.IsEmpty() {
		return
	}

	jobs := ctx.queue.Dequeue(100)
	var usernames []string
	for _, job := range jobs {
		usernames = append(usernames, job.Username)
	}

	profiles, err := usernamesToUuids(usernames)
	switch err.(type) {
	case *mojang.TooManyRequestsError:
		for _, job := range jobs {
			job.RespondTo <- nil
		}

		return
	case error:
		panic(err)
	}

	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go func(job *jobItem) {
			var result *mojang.SignedTexturesResponse
			shouldCache := true
			var uuid string
			for _, profile := range profiles {
				if strings.EqualFold(job.Username, profile.Name) {
					uuid = profile.Id
					break
				}
			}

			if uuid != "" {
				result, err = uuidToTextures(uuid, true)
				if err != nil {
					if _, ok := err.(*mojang.TooManyRequestsError); !ok {
						panic(err)
					}

					shouldCache = false
				}
			}

			wg.Done()

			job.RespondTo <- result

			if shouldCache {
				// TODO: store result to cache
			}
		}(job)
	}

	wg.Wait()
}
