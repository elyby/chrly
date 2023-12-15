package di

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/defval/di"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/api/mojang"
	es "github.com/elyby/chrly/eventsubscribers"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

var mojangTextures = di.Options(
	di.Invoke(interceptMojangApiUrls),
	di.Provide(newMojangTexturesProviderFactory),
	di.Provide(newMojangTexturesProvider),
	di.Provide(newMojangTexturesUuidsProviderFactory),
	di.Provide(newMojangTexturesBatchUUIDsProvider),
	di.Provide(newMojangTexturesBatchUUIDsProviderStrategyFactory),
	di.Provide(newMojangTexturesBatchUUIDsProviderDelayedStrategy),
	di.Provide(newMojangTexturesBatchUUIDsProviderFullBusStrategy),
	di.Provide(newMojangSignedTexturesProvider),
	di.Provide(newMojangTexturesStorageFactory),
)

func interceptMojangApiUrls(config *viper.Viper) error {
	apiUrl := config.GetString("mojang.api_base_url")
	if apiUrl != "" {
		u, err := url.ParseRequestURI(apiUrl)
		if err != nil {
			return err
		}

		mojang.ApiMojangDotComAddr = u.String()
	}

	sessionServerUrl := config.GetString("mojang.session_server_base_url")
	if sessionServerUrl != "" {
		u, err := url.ParseRequestURI(apiUrl)
		if err != nil {
			return err
		}

		mojang.SessionServerMojangComAddr = u.String()
	}

	return nil
}

func newMojangTexturesProviderFactory(
	container *di.Container,
	config *viper.Viper,
) (http.MojangTexturesProvider, error) {
	config.SetDefault("mojang_textures.enabled", true)
	if !config.GetBool("mojang_textures.enabled") {
		return &mojangtextures.NilProvider{}, nil
	}

	var provider *mojangtextures.Provider
	err := container.Resolve(&provider)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func newMojangTexturesProvider(
	emitter mojangtextures.Emitter,
	uuidsProvider mojangtextures.UUIDsProvider,
	texturesProvider mojangtextures.TexturesProvider,
	storage mojangtextures.Storage,
) *mojangtextures.Provider {
	return &mojangtextures.Provider{
		Emitter:          emitter,
		UUIDsProvider:    uuidsProvider,
		TexturesProvider: texturesProvider,
		Storage:          storage,
	}
}

func newMojangTexturesUuidsProviderFactory(
	container *di.Container,
) (mojangtextures.UUIDsProvider, error) {
	var provider *mojangtextures.BatchUuidsProvider
	err := container.Resolve(&provider)

	return provider, err
}

func newMojangTexturesBatchUUIDsProvider(
	container *di.Container,
	strategy mojangtextures.BatchUuidsProviderStrategy,
	emitter mojangtextures.Emitter,
) (*mojangtextures.BatchUuidsProvider, error) {
	if err := container.Provide(func(emitter es.Subscriber, config *viper.Viper) *namedHealthChecker {
		config.SetDefault("healthcheck.mojang_batch_uuids_provider_cool_down_duration", time.Minute)

		return &namedHealthChecker{
			Name: "mojang-batch-uuids-provider-response",
			Checker: es.MojangBatchUuidsProviderResponseChecker(
				emitter,
				config.GetDuration("healthcheck.mojang_batch_uuids_provider_cool_down_duration"),
			),
		}
	}); err != nil {
		return nil, err
	}

	if err := container.Provide(func(emitter es.Subscriber, config *viper.Viper) *namedHealthChecker {
		config.SetDefault("healthcheck.mojang_batch_uuids_provider_queue_length_limit", 50)

		return &namedHealthChecker{
			Name: "mojang-batch-uuids-provider-queue-length",
			Checker: es.MojangBatchUuidsProviderQueueLengthChecker(
				emitter,
				config.GetInt("healthcheck.mojang_batch_uuids_provider_queue_length_limit"),
			),
		}
	}); err != nil {
		return nil, err
	}

	return mojangtextures.NewBatchUuidsProvider(context.Background(), strategy, emitter), nil
}

func newMojangTexturesBatchUUIDsProviderStrategyFactory(
	container *di.Container,
	config *viper.Viper,
) (mojangtextures.BatchUuidsProviderStrategy, error) {
	config.SetDefault("queue.strategy", "periodic")

	strategyName := config.GetString("queue.strategy")
	switch strategyName {
	case "periodic":
		var strategy *mojangtextures.PeriodicStrategy
		err := container.Resolve(&strategy)
		if err != nil {
			return nil, err
		}

		return strategy, nil
	case "full-bus":
		var strategy *mojangtextures.FullBusStrategy
		err := container.Resolve(&strategy)
		if err != nil {
			return nil, err
		}

		return strategy, nil
	default:
		return nil, fmt.Errorf("unknown queue strategy \"%s\"", strategyName)
	}
}

func newMojangTexturesBatchUUIDsProviderDelayedStrategy(config *viper.Viper) *mojangtextures.PeriodicStrategy {
	config.SetDefault("queue.loop_delay", 2*time.Second+500*time.Millisecond)
	config.SetDefault("queue.batch_size", 10)

	return mojangtextures.NewPeriodicStrategy(
		config.GetDuration("queue.loop_delay"),
		config.GetInt("queue.batch_size"),
	)
}

func newMojangTexturesBatchUUIDsProviderFullBusStrategy(config *viper.Viper) *mojangtextures.FullBusStrategy {
	config.SetDefault("queue.loop_delay", 2*time.Second+500*time.Millisecond)
	config.SetDefault("queue.batch_size", 10)

	return mojangtextures.NewFullBusStrategy(
		config.GetDuration("queue.loop_delay"),
		config.GetInt("queue.batch_size"),
	)
}

func newMojangSignedTexturesProvider(emitter mojangtextures.Emitter) mojangtextures.TexturesProvider {
	return &mojangtextures.MojangApiTexturesProvider{
		Emitter: emitter,
	}
}

func newMojangTexturesStorageFactory(
	uuidsStorage mojangtextures.UUIDsStorage,
	texturesStorage mojangtextures.TexturesStorage,
) mojangtextures.Storage {
	return &mojangtextures.SeparatedStorage{
		UUIDsStorage:    uuidsStorage,
		TexturesStorage: texturesStorage,
	}
}
