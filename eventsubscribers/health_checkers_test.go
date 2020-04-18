package eventsubscribers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/dispatcher"
)

func TestMojangBatchUuidsProviderChecker(t *testing.T) {
	t.Run("empty state", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangBatchUuidsProviderResponseChecker(d, time.Millisecond)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("when no error occurred", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangBatchUuidsProviderResponseChecker(d, time.Millisecond)
		d.Emit("mojang_textures:batch_uuids_provider:result", []string{"username"}, []*mojang.ProfileInfo{}, nil)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("when error occurred", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangBatchUuidsProviderResponseChecker(d, time.Millisecond)
		err := errors.New("some error occurred")
		d.Emit("mojang_textures:batch_uuids_provider:result", []string{"username"}, nil, err)
		assert.Equal(t, err, checker(context.Background()))
	})

	t.Run("should reset value after passed duration", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangBatchUuidsProviderResponseChecker(d, 20*time.Millisecond)
		err := errors.New("some error occurred")
		d.Emit("mojang_textures:batch_uuids_provider:result", []string{"username"}, nil, err)
		assert.Equal(t, err, checker(context.Background()))
		time.Sleep(40 * time.Millisecond)
		assert.Nil(t, checker(context.Background()))
	})
}

func TestMojangBatchUuidsProviderQueueLengthChecker(t *testing.T) {
	t.Run("empty state", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangBatchUuidsProviderQueueLengthChecker(d, 10)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("less than allowed limit", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangBatchUuidsProviderQueueLengthChecker(d, 10)
		d.Emit("mojang_textures:batch_uuids_provider:round", []string{"username"}, 9)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("greater than allowed limit", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangBatchUuidsProviderQueueLengthChecker(d, 10)
		d.Emit("mojang_textures:batch_uuids_provider:round", []string{"username"}, 10)
		checkResult := checker(context.Background())
		if assert.Error(t, checkResult) {
			assert.Equal(t, "the maximum number of tasks in the queue has been exceeded", checkResult.Error())
		}
	})
}
