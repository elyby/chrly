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
	var mutex sync.Mutex
	var lastCallErr error
	var expireTimer *time.Timer
	dispatcher.Subscribe(
		"mojang_textures:batch_uuids_provider:result",
		func(usernames []string, profiles []*mojang.ProfileInfo, err error) {
			mutex.Lock()
			defer mutex.Unlock()

			lastCallErr = err
			if expireTimer != nil {
				expireTimer.Stop()
			}

			expireTimer = time.AfterFunc(resetDuration, func() {
				mutex.Lock()
				lastCallErr = nil
				mutex.Unlock()
			})
		},
	)

	return func(ctx context.Context) error {
		mutex.Lock()
		defer mutex.Unlock()

		return lastCallErr
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
