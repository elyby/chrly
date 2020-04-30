package eventsubscribers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/dispatcher"
)

type pingableMock struct {
	mock.Mock
}

func (p *pingableMock) Ping() error {
	args := p.Called()
	return args.Error(0)
}

func TestDatabaseChecker(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		p := &pingableMock{}
		p.On("Ping").Return(nil)
		checker := DatabaseChecker(p)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("mock error")
		p := &pingableMock{}
		p.On("Ping").Return(err)
		checker := DatabaseChecker(p)
		assert.Equal(t, err, checker(context.Background()))
	})

	t.Run("context timeout", func(t *testing.T) {
		p := &pingableMock{}
		waitChan := make(chan time.Time, 1)
		p.On("Ping").WaitUntil(waitChan).Return(nil)

		ctx, _ := context.WithTimeout(context.Background(), 0)
		checker := DatabaseChecker(p)
		assert.Errorf(t, checker(ctx), "check timeout")
		close(waitChan)
	})
}

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

func TestMojangApiTexturesProviderResponseChecker(t *testing.T) {
	t.Run("empty state", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangApiTexturesProviderResponseChecker(d, time.Millisecond)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("when no error occurred", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangApiTexturesProviderResponseChecker(d, time.Millisecond)
		d.Emit("mojang_textures:mojang_api_textures_provider:after_request",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			&mojang.SignedTexturesResponse{},
			nil,
		)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("when error occurred", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangApiTexturesProviderResponseChecker(d, time.Millisecond)
		err := errors.New("some error occurred")
		d.Emit("mojang_textures:mojang_api_textures_provider:after_request", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil, err)
		assert.Equal(t, err, checker(context.Background()))
	})

	t.Run("should reset value after passed duration", func(t *testing.T) {
		d := dispatcher.New()
		checker := MojangApiTexturesProviderResponseChecker(d, 20*time.Millisecond)
		err := errors.New("some error occurred")
		d.Emit("mojang_textures:mojang_api_textures_provider:after_request", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil, err)
		assert.Equal(t, err, checker(context.Background()))
		time.Sleep(40 * time.Millisecond)
		assert.Nil(t, checker(context.Background()))
	})
}
