package queue

import (
	"net"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mono83/slf/wd"

	"github.com/elyby/chrly/api/mojang"
)

var usernamesToUuids = mojang.UsernamesToUuids
var uuidToTextures = mojang.UuidToTextures
var uuidsQueueIterationDelay = time.Second
var forever = func() bool {
	return true
}

// https://help.mojang.com/customer/portal/articles/928638
var allowedUsernamesRegex = regexp.MustCompile(`^[\w_]{3,16}$`)

type JobsQueue struct {
	Storage Storage
	Logger  wd.Watchdog

	onFirstCall sync.Once
	queue       jobsQueue
	broadcast   *broadcastMap
}

func (ctx *JobsQueue) GetTexturesForUsername(username string) chan *mojang.SignedTexturesResponse {
	// TODO: convert username to lower case
	ctx.onFirstCall.Do(func() {
		ctx.queue.New()
		ctx.broadcast = newBroadcaster()
		ctx.startQueue()
	})

	responseChan := make(chan *mojang.SignedTexturesResponse)
	if !allowedUsernamesRegex.MatchString(username) {
		ctx.Logger.IncCounter("mojang_textures.invalid_username", 1)
		go func() {
			responseChan <- nil
			close(responseChan)
		}()

		return responseChan
	}

	ctx.Logger.IncCounter("mojang_textures.request", 1)

	uuid, err := ctx.Storage.GetUuid(username)
	if err == nil && uuid == "" {
		ctx.Logger.IncCounter("mojang_textures.usernames.cache_hit_nil", 1)

		go func() {
			responseChan <- nil
			close(responseChan)
		}()

		return responseChan
	}

	isFirstListener := ctx.broadcast.AddListener(username, responseChan)
	if isFirstListener {
		start := time.Now()
		// TODO: respond nil if processing takes more than 5 seconds

		resultChan := make(chan *mojang.SignedTexturesResponse)
		if uuid == "" {
			ctx.Logger.IncCounter("mojang_textures.usernames.queued", 1)
			ctx.queue.Enqueue(&jobItem{username, resultChan})
		} else {
			ctx.Logger.IncCounter("mojang_textures.usernames.cache_hit", 1)
			go func() {
				resultChan <- ctx.getTextures(uuid)
			}()
		}

		go func() {
			result := <-resultChan
			close(resultChan)
			ctx.broadcast.BroadcastAndRemove(username, result)
			ctx.Logger.RecordTimer("mojang_textures.result_time", time.Since(start))
		}()
	} else {
		ctx.Logger.IncCounter("mojang_textures.already_in_queue", 1)
	}

	return responseChan
}

func (ctx *JobsQueue) startQueue() {
	go func() {
		time.Sleep(uuidsQueueIterationDelay)
		for forever() {
			start := time.Now()
			ctx.queueRound()
			elapsed := time.Since(start)
			ctx.Logger.RecordTimer("mojang_textures.usernames.round_time", elapsed)
			time.Sleep(uuidsQueueIterationDelay)
		}
	}()
}

func (ctx *JobsQueue) queueRound() {
	if ctx.queue.IsEmpty() {
		return
	}

	queueSize := ctx.queue.Size()
	jobs := ctx.queue.Dequeue(10)
	ctx.Logger.UpdateGauge("mojang_textures.usernames.iteration_size", int64(len(jobs)))
	ctx.Logger.UpdateGauge("mojang_textures.usernames.queue_size", int64(queueSize-len(jobs)))
	var usernames []string
	for _, job := range jobs {
		usernames = append(usernames, job.Username)
	}

	profiles, err := usernamesToUuids(usernames)
	if err != nil {
		ctx.handleResponseError(err, "usernames")
		for _, job := range jobs {
			job.RespondTo <- nil
		}

		return
	}

	for _, job := range jobs {
		go func(job *jobItem) {
			var uuid string
			// The profiles in the response are not ordered, so we must search each username over full array
			for _, profile := range profiles {
				if strings.EqualFold(job.Username, profile.Name) {
					uuid = profile.Id
					break
				}
			}

			_ = ctx.Storage.StoreUuid(job.Username, uuid)
			if uuid == "" {
				job.RespondTo <- nil
				ctx.Logger.IncCounter("mojang_textures.usernames.uuid_miss", 1)
			} else {
				job.RespondTo <- ctx.getTextures(uuid)
				ctx.Logger.IncCounter("mojang_textures.usernames.uuid_hit", 1)
			}
		}(job)
	}
}

func (ctx *JobsQueue) getTextures(uuid string) *mojang.SignedTexturesResponse {
	existsTextures, err := ctx.Storage.GetTextures(uuid)
	if err == nil {
		ctx.Logger.IncCounter("mojang_textures.textures.cache_hit", 1)
		return existsTextures
	}

	ctx.Logger.IncCounter("mojang_textures.textures.request", 1)

	start := time.Now()
	result, err := uuidToTextures(uuid, true)
	ctx.Logger.RecordTimer("mojang_textures.textures.request_time", time.Since(start))
	if err != nil {
		ctx.handleResponseError(err, "textures")
	}

	// Mojang can respond with an error, but count it as a hit, so store result even if the textures is nil
	ctx.Storage.StoreTextures(uuid, result)

	return result
}

func (ctx *JobsQueue) handleResponseError(err error, threadName string) {
	ctx.Logger.Debug(":name: Got response error :err", wd.NameParam(threadName), wd.ErrParam(err))

	switch err.(type) {
	case mojang.ResponseError:
		if _, ok := err.(*mojang.BadRequestError); ok {
			ctx.Logger.Warning(":name: Got 400 Bad Request :err", wd.NameParam(threadName), wd.ErrParam(err))
			return
		}

		if _, ok := err.(*mojang.ForbiddenError); ok {
			ctx.Logger.Warning(":name: Got 403 Forbidden :err", wd.NameParam(threadName), wd.ErrParam(err))
			return
		}

		if _, ok := err.(*mojang.TooManyRequestsError); ok {
			ctx.Logger.Warning(":name: Got 429 Too Many Requests :err", wd.NameParam(threadName), wd.ErrParam(err))
			return
		}

		return
	case net.Error:
		if err.(net.Error).Timeout() {
			return
		}

		if _, ok := err.(*url.Error); ok {
			return
		}

		if opErr, ok := err.(*net.OpError); ok && (opErr.Op == "dial" || opErr.Op == "read") {
			return
		}

		if err == syscall.ECONNREFUSED {
			return
		}
	}

	ctx.Logger.Emergency(":name: Unknown Mojang response error: :err", wd.NameParam(threadName), wd.ErrParam(err))
}
