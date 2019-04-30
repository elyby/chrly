package queue

import (
	"github.com/elyby/chrly/api/mojang"

	testify "github.com/stretchr/testify/assert"
	"testing"
)

func TestBroadcastMap_GetOrAppend(t *testing.T) {
	t.Run("first call when username didn't exist before should return true", func(t *testing.T) {
		assert := testify.New(t)

		broadcaster := newBroadcaster()
		channel := make(chan *mojang.SignedTexturesResponse)
		isFirstListener := broadcaster.AddListener("mock", channel)

		assert.True(isFirstListener)
		listeners, ok := broadcaster.listeners["mock"]
		assert.True(ok)
		assert.Len(listeners, 1)
		assert.Equal(channel, listeners[0])
	})

	t.Run("subsequent calls should return false", func(t *testing.T) {
		assert := testify.New(t)

		broadcaster := newBroadcaster()
		channel1 := make(chan *mojang.SignedTexturesResponse)
		isFirstListener := broadcaster.AddListener("mock", channel1)

		assert.True(isFirstListener)

		channel2 := make(chan *mojang.SignedTexturesResponse)
		isFirstListener = broadcaster.AddListener("mock", channel2)

		assert.False(isFirstListener)

		channel3 := make(chan *mojang.SignedTexturesResponse)
		isFirstListener = broadcaster.AddListener("mock", channel3)

		assert.False(isFirstListener)
	})
}

func TestBroadcastMap_BroadcastAndRemove(t *testing.T) {
	t.Run("should broadcast to all listeners and remove the key", func(t *testing.T) {
		assert := testify.New(t)

		broadcaster := newBroadcaster()
		channel1 := make(chan *mojang.SignedTexturesResponse)
		channel2 := make(chan *mojang.SignedTexturesResponse)
		broadcaster.AddListener("mock", channel1)
		broadcaster.AddListener("mock", channel2)

		result := &mojang.SignedTexturesResponse{Id: "mockUuid"}
		broadcaster.BroadcastAndRemove("mock", result)

		assert.Equal(result, <-channel1)
		assert.Equal(result, <-channel2)

		channel3 := make(chan *mojang.SignedTexturesResponse)
		isFirstListener := broadcaster.AddListener("mock", channel3)
		assert.True(isFirstListener)
	})

	t.Run("call on not exists username", func(t *testing.T) {
		assert := testify.New(t)

		assert.NotPanics(func() {
			broadcaster := newBroadcaster()
			broadcaster.BroadcastAndRemove("mock", &mojang.SignedTexturesResponse{})
		})
	})
}
