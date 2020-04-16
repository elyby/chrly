package di

import (
	"fmt"
	"net/url"

	"github.com/goava/di"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

var mojangTextures = di.Options(
	di.Provide(newMojangTexturesProviderFactory),
	di.Provide(newMojangTexturesProvider),
	di.Provide(newMojangTexturesUuidsProvider),
	di.Provide(newMojangSignedTexturesProvider),
	di.Provide(newMojangTexturesStorageFactory),
)

func newMojangTexturesProviderFactory(
	container *di.Container,
	config *viper.Viper,
) (http.MojangTexturesProvider, error) {
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

func newMojangTexturesUuidsProvider(
	config *viper.Viper,
	emitter mojangtextures.Emitter,
) (mojangtextures.UUIDsProvider, error) {
	preferredUuidsProvider := config.GetString("mojang_textures.uuids_provider.driver")
	if preferredUuidsProvider == "remote" {
		remoteUrl, err := url.Parse(config.GetString("mojang_textures.uuids_provider.url"))
		if err != nil {
			return nil, fmt.Errorf("Unable to parse remote url: %w", err)
		}

		return &mojangtextures.RemoteApiUuidsProvider{
			Emitter: emitter,
			Url:     *remoteUrl,
		}, nil
	}

	return &mojangtextures.BatchUuidsProvider{
		Emitter:        emitter,
		IterationDelay: config.GetDuration("queue.loop_delay"),
		IterationSize:  config.GetInt("queue.batch_size"),
	}, nil
}

func newMojangSignedTexturesProvider(emitter mojangtextures.Emitter) mojangtextures.TexturesProvider {
	return &mojangtextures.MojangApiTexturesProvider{
		Emitter: emitter,
	}
}

func newMojangTexturesStorageFactory(
	uuidsStorage mojangtextures.UuidsStorage,
	texturesStorage mojangtextures.TexturesStorage,
) mojangtextures.Storage {
	return &mojangtextures.SeparatedStorage{
		UuidsStorage:    uuidsStorage,
		TexturesStorage: texturesStorage,
	}
}
