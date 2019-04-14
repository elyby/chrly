package queue

import (
	"strings"
	"sync"
	"time"

	"github.com/elyby/chrly/api/mojang"
)

var once sync.Once
var jobsQueue = JobsQueue{}

func ScheduleTexturesForUsername(username string) chan *mojang.SignedTexturesResponse {
	once.Do(func() {
		jobsQueue.New()
		startQueue()
	})

	// TODO: prevent of adding the same username more than once
	resultChan := make(chan *mojang.SignedTexturesResponse)
	jobsQueue.Enqueue(&Job{username, resultChan})

	return resultChan
}

func startQueue() {
	go func() {
		for {
			start := time.Now()
			queueRound()
			time.Sleep(time.Second - time.Since(start))
		}
	}()
}

func queueRound() {
	if jobsQueue.IsEmpty() {
		return
	}

	jobs := jobsQueue.Dequeue(100)
	var usernames []string
	for _, job := range jobs {
		usernames = append(usernames, job.Username)
	}

	profiles, err := mojang.UsernamesToUuids(usernames)
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
		go func() {
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
				result, err = mojang.UuidToTextures(uuid, true)
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
		}()
	}

	wg.Wait()
}
