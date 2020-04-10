package eventsubscribers

import (
	"context"
	"sync"

	"github.com/etherlabsio/healthcheck"

	"github.com/elyby/chrly/api/mojang"
)

func MojangBatchUuidsProviderChecker(dispatcher Subscriber) healthcheck.CheckerFunc {
	var mutex sync.Mutex
	var lastCallErr error // TODO: need to reset this value after some time
	dispatcher.Subscribe(
		"mojang_textures:batch_uuids_provider:result",
		func(usernames []string, profiles []*mojang.ProfileInfo, err error) {
			mutex.Lock()
			lastCallErr = err
			mutex.Unlock()
		},
	)

	return func(ctx context.Context) error {
		mutex.Lock()
		defer mutex.Unlock()

		return lastCallErr
	}
}
