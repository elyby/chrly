package eventsubscribers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/bootstrap"
)

func TestMojangBatchUuidsProviderChecker(t *testing.T) {
	t.Run("empty state", func(t *testing.T) {
		dispatcher := bootstrap.CreateEventDispatcher()
		checker := MojangBatchUuidsProviderChecker(dispatcher, time.Millisecond)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("when no error occurred", func(t *testing.T) {
		dispatcher := bootstrap.CreateEventDispatcher()
		checker := MojangBatchUuidsProviderChecker(dispatcher, time.Millisecond)
		dispatcher.Emit("mojang_textures:batch_uuids_provider:result", []string{"username"}, []*mojang.ProfileInfo{}, nil)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("when error occurred", func(t *testing.T) {
		dispatcher := bootstrap.CreateEventDispatcher()
		checker := MojangBatchUuidsProviderChecker(dispatcher, time.Millisecond)
		err := errors.New("some error occurred")
		dispatcher.Emit("mojang_textures:batch_uuids_provider:result", []string{"username"}, nil, err)
		assert.Equal(t, err, checker(context.Background()))
	})

	t.Run("should reset value after passed duration", func(t *testing.T) {
		dispatcher := bootstrap.CreateEventDispatcher()
		checker := MojangBatchUuidsProviderChecker(dispatcher, 20*time.Millisecond)
		err := errors.New("some error occurred")
		dispatcher.Emit("mojang_textures:batch_uuids_provider:result", []string{"username"}, nil, err)
		assert.Equal(t, err, checker(context.Background()))
		time.Sleep(40 * time.Millisecond)
		assert.Nil(t, checker(context.Background()))
	})
}
