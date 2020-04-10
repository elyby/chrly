package eventsubscribers

import (
	"context"
	"sync"
	"time"

	"github.com/etherlabsio/healthcheck"

	"github.com/elyby/chrly/api/mojang"
)

func MojangBatchUuidsProviderChecker(dispatcher Subscriber, resetDuration time.Duration) healthcheck.CheckerFunc {
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
