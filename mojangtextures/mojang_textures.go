package mojangtextures

import (
	"errors"
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

type broadcastResult struct {
	textures *mojang.SignedTexturesResponse
	error    error
}

type broadcaster struct {
	lock      sync.Mutex
	listeners map[string][]chan *broadcastResult
}

func createBroadcaster() *broadcaster {
	return &broadcaster{
		listeners: make(map[string][]chan *broadcastResult),
	}
}

// Returns a boolean value, which will be true if the passed username didn't exist before
func (c *broadcaster) AddListener(username string, resultChan chan *broadcastResult) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, alreadyHasSource := c.listeners[username]
	if alreadyHasSource {
		c.listeners[username] = append(val, resultChan)
		return false
	}

	c.listeners[username] = []chan *broadcastResult{resultChan}

	return true
}

func (c *broadcaster) BroadcastAndRemove(username string, result *broadcastResult) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.listeners[username]
	if !ok {
		return
	}

	for _, channel := range val {
		go func(channel chan *broadcastResult) {
			channel <- result
			close(channel)
		}(channel)
	}

	delete(c.listeners, username)
}

// https://help.mojang.com/customer/portal/articles/928638
var allowedUsernamesRegex = regexp.MustCompile(`^[\w_]{3,16}$`)

type UuidsProvider interface {
	GetUuid(username string) (*mojang.ProfileInfo, error)
}

type TexturesProvider interface {
	GetTextures(uuid string) (*mojang.SignedTexturesResponse, error)
}

type Provider struct {
	UuidsProvider
	TexturesProvider
	Storage
	Logger wd.Watchdog

	onFirstCall sync.Once
	*broadcaster
}

func (ctx *Provider) GetForUsername(username string) (*mojang.SignedTexturesResponse, error) {
	ctx.onFirstCall.Do(func() {
		ctx.broadcaster = createBroadcaster()
	})

	if !allowedUsernamesRegex.MatchString(username) {
		ctx.Logger.IncCounter("mojang_textures.invalid_username", 1)
		return nil, errors.New("invalid username")
	}

	username = strings.ToLower(username)
	ctx.Logger.IncCounter("mojang_textures.request", 1)

	uuid, err := ctx.Storage.GetUuid(username)
	if err == nil && uuid == "" {
		ctx.Logger.IncCounter("mojang_textures.usernames.cache_hit_nil", 1)
		return nil, nil
	}

	if uuid != "" {
		ctx.Logger.IncCounter("mojang_textures.usernames.cache_hit", 1)
		textures, err := ctx.Storage.GetTextures(uuid)
		if err == nil {
			ctx.Logger.IncCounter("mojang_textures.textures.cache_hit", 1)
			return textures, nil
		}
	}

	resultChan := make(chan *broadcastResult)
	isFirstListener := ctx.broadcaster.AddListener(username, resultChan)
	if isFirstListener {
		go ctx.getResultAndBroadcast(username, uuid)
	} else {
		ctx.Logger.IncCounter("mojang_textures.already_scheduled", 1)
	}

	result := <-resultChan

	return result.textures, result.error
}

func (ctx *Provider) getResultAndBroadcast(username string, uuid string) {
	start := time.Now()

	result := ctx.getResult(username, uuid)
	ctx.broadcaster.BroadcastAndRemove(username, result)

	ctx.Logger.RecordTimer("mojang_textures.result_time", time.Since(start))
}

func (ctx *Provider) getResult(username string, uuid string) *broadcastResult {
	if uuid == "" {
		profile, err := ctx.UuidsProvider.GetUuid(username)
		if err != nil {
			ctx.handleMojangApiResponseError(err, "usernames")
			return &broadcastResult{nil, err}
		}

		uuid = ""
		if profile != nil {
			uuid = profile.Id
		}

		_ = ctx.Storage.StoreUuid(username, uuid)

		if uuid == "" {
			ctx.Logger.IncCounter("mojang_textures.usernames.uuid_miss", 1)
			return &broadcastResult{nil, nil}
		}

		ctx.Logger.IncCounter("mojang_textures.usernames.uuid_hit", 1)
	}

	textures, err := ctx.TexturesProvider.GetTextures(uuid)
	if err != nil {
		ctx.handleMojangApiResponseError(err, "textures")
		return &broadcastResult{nil, err}
	}

	// Mojang can respond with an error, but it will still count as a hit,
	// therefore store the result even if textures is nil to prevent 429 error
	ctx.Storage.StoreTextures(uuid, textures)

	if textures != nil {
		ctx.Logger.IncCounter("mojang_textures.usernames.textures_hit", 1)
	} else {
		ctx.Logger.IncCounter("mojang_textures.usernames.textures_miss", 1)
	}

	return &broadcastResult{textures, nil}
}

func (ctx *Provider) handleMojangApiResponseError(err error, threadName string) {
	errParam := wd.ErrParam(err)
	threadParam := wd.NameParam(threadName)

	ctx.Logger.Debug(":name: Got response error :err", threadParam, errParam)

	switch err.(type) {
	case mojang.ResponseError:
		if _, ok := err.(*mojang.BadRequestError); ok {
			ctx.Logger.Warning(":name: Got 400 Bad Request :err", threadParam, errParam)
			return
		}

		if _, ok := err.(*mojang.ForbiddenError); ok {
			ctx.Logger.Warning(":name: Got 403 Forbidden :err", threadParam, errParam)
			return
		}

		if _, ok := err.(*mojang.TooManyRequestsError); ok {
			ctx.Logger.Warning(":name: Got 429 Too Many Requests :err", threadParam, errParam)
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

	ctx.Logger.Emergency(":name: Unknown Mojang response error: :err", threadParam, errParam)
}
