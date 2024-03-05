package di

import (
	"net/http"
	"net/url"
	"time"

	"github.com/defval/di"
	"github.com/spf13/viper"

	"ely.by/chrly/internal/mojang"
	"ely.by/chrly/internal/profiles"
)

var mojangDiOptions = di.Options(
	di.Provide(newMojangApi),
	di.Provide(newMojangTexturesProviderFactory),
	di.Provide(newMojangTexturesProvider),
	di.Provide(newMojangTexturesUuidsProviderFactory),
	di.Provide(newMojangTexturesBatchUUIDsProvider),
	di.Provide(newMojangSignedTexturesProvider),
)

func newMojangApi(config *viper.Viper, httpClient *http.Client) (*mojang.MojangApi, error) {
	batchUuidsUrl := config.GetString("mojang.batch_uuids_url")
	if batchUuidsUrl != "" {
		if _, err := url.ParseRequestURI(batchUuidsUrl); err != nil {
			return nil, err
		}
	}

	profileUrl := config.GetString("mojang.profile_url")
	if profileUrl != "" {
		if _, err := url.ParseRequestURI(batchUuidsUrl); err != nil {
			return nil, err
		}
	}

	return mojang.NewMojangApi(httpClient, batchUuidsUrl, profileUrl), nil
}

func newMojangTexturesProviderFactory(
	container *di.Container,
	config *viper.Viper,
) (profiles.MojangProfilesProvider, error) {
	config.SetDefault("mojang_textures.enabled", true)
	if !config.GetBool("mojang_textures.enabled") {
		return &mojang.NilProvider{}, nil
	}

	var provider *mojang.MojangTexturesProvider
	err := container.Resolve(&provider)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func newMojangTexturesProvider(
	uuidsProvider mojang.UuidsProvider,
	texturesProvider mojang.TexturesProvider,
) (*mojang.MojangTexturesProvider, error) {
	return mojang.NewMojangTexturesProvider(
		uuidsProvider,
		texturesProvider,
	)
}

func newMojangTexturesUuidsProviderFactory(
	batchProvider *mojang.BatchUuidsProvider,
	uuidsStorage mojang.MojangUuidsStorage,
) (mojang.UuidsProvider, error) {
	return mojang.NewUuidsProviderWithCache(batchProvider, uuidsStorage)
}

func newMojangTexturesBatchUUIDsProvider(
	mojangApi *mojang.MojangApi,
	config *viper.Viper,
) (*mojang.BatchUuidsProvider, error) {
	config.SetDefault("queue.loop_delay", 2*time.Second+500*time.Millisecond)
	config.SetDefault("queue.batch_size", 10)
	config.SetDefault("queue.strategy", "periodic")

	return mojang.NewBatchUuidsProvider(
		mojangApi.UsernamesToUuids,
		config.GetInt("queue.batch_size"),
		config.GetDuration("queue.loop_delay"),
		config.GetString("queue.strategy") == "full-bus",
	)
}

func newMojangSignedTexturesProvider(mojangApi *mojang.MojangApi) (mojang.TexturesProvider, error) {
	provider, err := mojang.NewMojangApiTexturesProvider(mojangApi.UuidToTextures)
	if err != nil {
		return nil, err
	}

	return mojang.NewTexturesProviderWithInMemoryCache(provider)
}
