package eventsubscribers

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/etherlabsio/healthcheck"

	"github.com/elyby/chrly/api/mojang"
)

type Pingable interface {
	Ping() error
}

func DatabaseChecker(connection Pingable) healthcheck.CheckerFunc {
	return func(ctx context.Context) error {
		done := make(chan error)
		go func() {
			done <- connection.Ping()
		}()

		select {
		case <-ctx.Done():
			return errors.New("check timeout")
		case err := <-done:
			return err
		}
	}
}

func MojangBatchUuidsProviderResponseChecker(dispatcher Subscriber, resetDuration time.Duration) healthcheck.CheckerFunc {
	errHolder := &expiringErrHolder{D: resetDuration}
	dispatcher.Subscribe(
		"mojang_textures:batch_uuids_provider:result",
		func(usernames []string, profiles []*mojang.ProfileInfo, err error) {
			errHolder.Set(err)
		},
	)

	return func(ctx context.Context) error {
		return errHolder.Get()
	}
}

func MojangBatchUuidsProviderQueueLengthChecker(dispatcher Subscriber, maxLength int) healthcheck.CheckerFunc {
	var mutex sync.Mutex
	queueLength := 0
	dispatcher.Subscribe("mojang_textures:batch_uuids_provider:round", func(usernames []string, tasksInQueue int) {
		mutex.Lock()
		queueLength = tasksInQueue
		mutex.Unlock()
	})

	return func(ctx context.Context) error {
		mutex.Lock()
		defer mutex.Unlock()

		if queueLength < maxLength {
			return nil
		}

		return errors.New("the maximum number of tasks in the queue has been exceeded")
	}
}

func MojangApiTexturesProviderResponseChecker(dispatcher Subscriber, resetDuration time.Duration) healthcheck.CheckerFunc {
	errHolder := &expiringErrHolder{D: resetDuration}
	dispatcher.Subscribe(
		"mojang_textures:mojang_api_textures_provider:after_request",
		func(uuid string, profile *mojang.SignedTexturesResponse, err error) {
			errHolder.Set(err)
		},
	)

	return func(ctx context.Context) error {
		return errHolder.Get()
	}
}

type expiringErrHolder struct {
	D   time.Duration
	err error
	l   sync.Mutex
	t   *time.Timer
}

func (h *expiringErrHolder) Get() error {
	h.l.Lock()
	defer h.l.Unlock()

	return h.err
}

func (h *expiringErrHolder) Set(err error) {
	h.l.Lock()
	defer h.l.Unlock()
	if h.t != nil {
		h.t.Stop()
		h.t = nil
	}

	h.err = err
	if err != nil {
		h.t = time.AfterFunc(h.D, func() {
			h.Set(nil)
		})
	}
}
