package queue

import (
	"sync"

	"github.com/elyby/chrly/api/mojang"
)

type broadcastMap struct {
	lock      sync.Mutex
	listeners map[string][]chan *mojang.SignedTexturesResponse
}

func newBroadcaster() *broadcastMap {
	return &broadcastMap{
		listeners: make(map[string][]chan *mojang.SignedTexturesResponse),
	}
}

// Returns a boolean value, which will be true if the username passed didn't exist before
func (c *broadcastMap) AddListener(username string, resultChan chan *mojang.SignedTexturesResponse) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, alreadyHasSource := c.listeners[username]
	if alreadyHasSource {
		c.listeners[username] = append(val, resultChan)
		return false
	}

	c.listeners[username] = []chan *mojang.SignedTexturesResponse{resultChan}

	return true
}

func (c *broadcastMap) BroadcastAndRemove(username string, result *mojang.SignedTexturesResponse) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.listeners[username]
	if !ok {
		return
	}

	for _, channel := range val {
		go func(channel chan *mojang.SignedTexturesResponse) {
			channel <- result
			close(channel)
		}(channel)
	}

	delete(c.listeners, username)
}
