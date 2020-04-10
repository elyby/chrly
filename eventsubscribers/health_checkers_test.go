package eventsubscribers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/bootstrap"
)

func TestMojangBatchUuidsProviderChecker(t *testing.T) {
	t.Run("empty state", func(t *testing.T) {
		dispatcher := bootstrap.CreateEventDispatcher()
		checker := MojangBatchUuidsProviderChecker(dispatcher)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("when no error occurred", func(t *testing.T) {
		dispatcher := bootstrap.CreateEventDispatcher()
		checker := MojangBatchUuidsProviderChecker(dispatcher)
		dispatcher.Emit("mojang_textures:batch_uuids_provider:result", []string{"username"}, []*mojang.ProfileInfo{}, nil)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("when error occurred", func(t *testing.T) {
		dispatcher := bootstrap.CreateEventDispatcher()
		checker := MojangBatchUuidsProviderChecker(dispatcher)
		err := errors.New("some error occurred")
		dispatcher.Emit("mojang_textures:batch_uuids_provider:result", []string{"username"}, nil, err)
		assert.Equal(t, err, checker(context.Background()))
	})
}
