package queue

import (
	"regexp"
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

// https://help.mojang.com/customer/portal/articles/928638
var allowedUsernamesRegex = regexp.MustCompile(`^[\w_]{3,16}$`)

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
	if !allowedUsernamesRegex.MatchString(username) {
		go func() {
			responseChan <- nil
			close(responseChan)
		}()

		return responseChan
	}

	uuid, err := ctx.Storage.GetUuid(username)
	if err == nil && uuid == "" {
		go func() {
			responseChan <- nil
			close(responseChan)
		}()

		return responseChan
	}

	isFirstListener := ctx.broadcast.AddListener(username, responseChan)
	if isFirstListener {
		// TODO: respond nil if processing takes more than 5 seconds

		resultChan := make(chan *mojang.SignedTexturesResponse)
		if uuid == "" {
			ctx.queue.Enqueue(&jobItem{username, resultChan})
		} else {
			go func() {
				resultChan <- ctx.getTextures(uuid)
			}()
		}

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
	case *mojang.TooManyRequestsError, *mojang.ServerError:
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
			var uuid string
			// Profiles in response not ordered, so we must search each username over full array
			for _, profile := range profiles {
				if strings.EqualFold(job.Username, profile.Name) {
					uuid = profile.Id
					break
				}
			}

			ctx.Storage.StoreUuid(job.Username, uuid)
			if uuid == "" {
				job.RespondTo <- nil
			} else {
				job.RespondTo <- ctx.getTextures(uuid)
			}

			wg.Done()
		}(job)
	}

	wg.Wait()
}

func (ctx *JobsQueue) getTextures(uuid string) *mojang.SignedTexturesResponse {
	existsTextures, err := ctx.Storage.GetTextures(uuid)
	if err == nil {
		return existsTextures
	}

	shouldCache := true
	result, err := uuidToTextures(uuid, true)
	switch err.(type) {
	case *mojang.EmptyResponse, *mojang.TooManyRequestsError, *mojang.ServerError:
		shouldCache = false
	case error:
		panic(err)
	}

	if shouldCache && result != nil {
		ctx.Storage.StoreTextures(result)
	}

	return result
}
