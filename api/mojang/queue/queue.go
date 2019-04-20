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
var forever = func() bool {
	return true
}

type JobsQueue struct {
	Storage Storage

	onFirstCall sync.Once
	queue       jobsQueue
	broadcast   *broadcastMap
}

func (ctx *JobsQueue) GetTexturesForUsername(username string) chan *mojang.SignedTexturesResponse {
	ctx.onFirstCall.Do(func() {
		ctx.queue.New()
		ctx.broadcast = newBroadcaster()
		ctx.startQueue()
	})

	responseChan := make(chan *mojang.SignedTexturesResponse)

	cachedResult := ctx.Storage.Get(username)
	if cachedResult != nil {
		go func() {
			responseChan <- cachedResult
			close(responseChan)
		}()

		return responseChan
	}

	isFirstListener := ctx.broadcast.AddListener(username, responseChan)
	if isFirstListener {
		resultChan := make(chan *mojang.SignedTexturesResponse)
		ctx.queue.Enqueue(&jobItem{username, resultChan})
		// TODO: return nil if processing takes more than 5 seconds

		go func() {
			result := <-resultChan
			close(resultChan)
			ctx.broadcast.BroadcastAndRemove(username, result)
		}()
	}

	return responseChan
}

func (ctx *JobsQueue) startQueue() {
	go func() {
		time.Sleep(delay)
		for forever() {
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
				var err error
				result, err = uuidToTextures(uuid, true)
				if err != nil {
					if _, ok := err.(*mojang.TooManyRequestsError); !ok {
						panic(err)
					}

					shouldCache = false
				}
			}

			wg.Done()

			if shouldCache && result != nil {
				ctx.Storage.Set(result)
			}

			job.RespondTo <- result
		}(job)
	}

	wg.Wait()
}
