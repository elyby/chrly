package mojangtextures

import (
	"errors"
	"regexp"
	"strings"
	"sync"

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

type UUIDsProvider interface {
	GetUuid(username string) (*mojang.ProfileInfo, error)
}

type TexturesProvider interface {
	GetTextures(uuid string) (*mojang.SignedTexturesResponse, error)
}

type Emitter interface {
	Emit(name string, args ...interface{})
}

type Provider struct {
	Emitter
	UUIDsProvider
	TexturesProvider
	Storage

	onFirstCall sync.Once
	*broadcaster
}

// TODO: move cache events on the corresponding level

func (ctx *Provider) GetForUsername(username string) (*mojang.SignedTexturesResponse, error) {
	ctx.onFirstCall.Do(func() {
		ctx.broadcaster = createBroadcaster()
	})

	if !allowedUsernamesRegex.MatchString(username) {
		return nil, errors.New("invalid username")
	}

	username = strings.ToLower(username)
	ctx.Emit("mojang_textures:call")

	uuid, err := ctx.Storage.GetUuid(username)
	if err == nil && uuid == "" {
		ctx.Emit("mojang_textures:usernames:cache_hit_nil")
		return nil, nil
	}

	if uuid != "" {
		ctx.Emit("mojang_textures:usernames:cache_hit")
		textures, err := ctx.Storage.GetTextures(uuid)
		if err == nil {
			ctx.Emit("mojang_textures:textures:cache_hit")
			return textures, nil
		}
	}

	resultChan := make(chan *broadcastResult)
	isFirstListener := ctx.broadcaster.AddListener(username, resultChan)
	if isFirstListener {
		go ctx.getResultAndBroadcast(username, uuid)
	} else {
		ctx.Emit("mojang_textures:already_processing")
	}

	result := <-resultChan

	return result.textures, result.error
}

func (ctx *Provider) getResultAndBroadcast(username string, uuid string) {
	ctx.Emit("mojang_textures:before_get_result")

	result := ctx.getResult(username, uuid)
	ctx.broadcaster.BroadcastAndRemove(username, result)

	ctx.Emit("mojang_textures:after_get_result")
}

func (ctx *Provider) getResult(username string, uuid string) *broadcastResult {
	if uuid == "" {
		profile, err := ctx.UUIDsProvider.GetUuid(username)
		if err != nil {
			ctx.Emit("mojang_textures:usernames:error", err)
			return &broadcastResult{nil, err}
		}

		uuid = ""
		if profile != nil {
			uuid = profile.Id
		}

		_ = ctx.Storage.StoreUuid(username, uuid)

		if uuid == "" {
			ctx.Emit("mojang_textures:usernames:uuid_miss")
			return &broadcastResult{nil, nil}
		}

		ctx.Emit("mojang_textures:usernames:uuid_hit")
	}

	textures, err := ctx.TexturesProvider.GetTextures(uuid)
	if err != nil {
		ctx.Emit("mojang_textures:textures:error", err)
		return &broadcastResult{nil, err}
	}

	// Mojang can respond with an error, but it will still count as a hit,
	// therefore store the result even if textures is nil to prevent 429 error
	ctx.Storage.StoreTextures(uuid, textures)

	if textures != nil {
		ctx.Emit("mojang_textures:textures:hit")
	} else {
		ctx.Emit("mojang_textures:textures:miss")
	}

	return &broadcastResult{textures, nil}
}
