package mojangtextures

import (
	"regexp"
	"strings"
	"sync"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/dispatcher"
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

// https://help.minecraft.net/hc/en-us/articles/4408950195341#h_01GE5JX1Z0CZ833A7S54Y195KV
var allowedUsernamesRegex = regexp.MustCompile(`(?i)^[0-9a-z_]{3,16}$`)

type UUIDsProvider interface {
	GetUuid(username string) (*mojang.ProfileInfo, error)
}

type TexturesProvider interface {
	GetTextures(uuid string) (*mojang.SignedTexturesResponse, error)
}

type Emitter interface {
	dispatcher.Emitter
}

type Provider struct {
	Emitter
	UUIDsProvider
	TexturesProvider
	Storage

	onFirstCall sync.Once
	*broadcaster
}

func (ctx *Provider) GetForUsername(username string) (*mojang.SignedTexturesResponse, error) {
	ctx.onFirstCall.Do(func() {
		ctx.broadcaster = createBroadcaster()
	})

	if !allowedUsernamesRegex.MatchString(username) {
		return nil, nil
	}

	username = strings.ToLower(username)
	ctx.Emit("mojang_textures:call", username)

	uuid, found, err := ctx.getUuidFromCache(username)
	if err != nil {
		return nil, err
	}

	if found && uuid == "" {
		return nil, nil
	}

	if uuid != "" {
		textures, err := ctx.getTexturesFromCache(uuid)
		if err == nil && textures != nil {
			return textures, nil
		}
	}

	resultChan := make(chan *broadcastResult)
	isFirstListener := ctx.broadcaster.AddListener(username, resultChan)
	if isFirstListener {
		go ctx.getResultAndBroadcast(username, uuid)
	} else {
		ctx.Emit("mojang_textures:already_processing", username)
	}

	result := <-resultChan

	return result.textures, result.error
}

func (ctx *Provider) getResultAndBroadcast(username string, uuid string) {
	ctx.Emit("mojang_textures:before_result", username, uuid)
	result := ctx.getResult(username, uuid)
	ctx.Emit("mojang_textures:after_result", username, result.textures, result.error)

	ctx.broadcaster.BroadcastAndRemove(username, result)
}

func (ctx *Provider) getResult(username string, cachedUuid string) *broadcastResult {
	uuid := cachedUuid
	if uuid == "" {
		profile, err := ctx.getUuid(username)
		if err != nil {
			return &broadcastResult{nil, err}
		}

		uuid = ""
		if profile != nil {
			uuid = profile.Id
		}

		_ = ctx.Storage.StoreUuid(username, uuid)

		if uuid == "" {
			return &broadcastResult{nil, nil}
		}
	}

	textures, err := ctx.getTextures(uuid)
	if err != nil {
		// Previously cached UUIDs may disappear
		// In this case we must invalidate UUID cache for given username
		if _, ok := err.(*mojang.EmptyResponse); ok && cachedUuid != "" {
			return ctx.getResult(username, "")
		}

		return &broadcastResult{nil, err}
	}

	// Mojang can respond with an error, but it will still count as a hit,
	// therefore store the result even if textures is nil to prevent 429 error
	ctx.Storage.StoreTextures(uuid, textures)

	return &broadcastResult{textures, nil}
}

func (ctx *Provider) getUuidFromCache(username string) (string, bool, error) {
	ctx.Emit("mojang_textures:usernames:before_cache", username)
	uuid, found, err := ctx.Storage.GetUuid(username)
	ctx.Emit("mojang_textures:usernames:after_cache", username, uuid, found, err)

	return uuid, found, err
}

func (ctx *Provider) getTexturesFromCache(uuid string) (*mojang.SignedTexturesResponse, error) {
	ctx.Emit("mojang_textures:textures:before_cache", uuid)
	textures, err := ctx.Storage.GetTextures(uuid)
	ctx.Emit("mojang_textures:textures:after_cache", uuid, textures, err)

	return textures, err
}

func (ctx *Provider) getUuid(username string) (*mojang.ProfileInfo, error) {
	ctx.Emit("mojang_textures:usernames:before_call", username)
	profile, err := ctx.UUIDsProvider.GetUuid(username)
	ctx.Emit("mojang_textures:usernames:after_call", username, profile, err)

	return profile, err
}

func (ctx *Provider) getTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	ctx.Emit("mojang_textures:textures:before_call", uuid)
	textures, err := ctx.TexturesProvider.GetTextures(uuid)
	ctx.Emit("mojang_textures:textures:after_call", uuid, textures, err)

	return textures, err
}
