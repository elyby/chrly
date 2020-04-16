package di

import (
	"github.com/goava/di"
	"github.com/spf13/viper"

	dbModule "github.com/elyby/chrly/db"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

var db = di.Options(
	di.Provide(newRedisFactory, di.WithName("redis")),
	di.Provide(newFSFactory, di.WithName("fs")),
	di.Provide(newSkinsRepository),
	di.Provide(newCapesRepository),
	di.Provide(newMojangUUIDsRepository),
	di.Provide(newMojangSignedTexturesStorage),
)

func newRedisFactory(config *viper.Viper) dbModule.RepositoriesCreator {
	return &dbModule.RedisFactory{
		Host:     config.GetString("storage.redis.host"),
		Port:     config.GetInt("storage.redis.port"),
		PoolSize: config.GetInt("storage.redis.poolSize"),
	}
}

func newFSFactory(config *viper.Viper) dbModule.RepositoriesCreator {
	return &dbModule.FilesystemFactory{
		BasePath:     config.GetString("storage.filesystem.basePath"),
		CapesDirName: config.GetString("storage.filesystem.capesDirName"),
	}
}

// v4 had the idea that it would be possible to separate backends for storing skins and capes.
// But in v5 the storage will be unified, so this is just temporary constructors before large reworking.
//
// Since there are no options for selecting target backends,
// all constants in this case point to static specific implementations.

func newSkinsRepository(container *di.Container) (http.SkinsRepository, error) {
	var factory dbModule.RepositoriesCreator
	err := container.Resolve(&factory, di.Name("redis"))
	if err != nil {
		return nil, err
	}

	return factory.CreateSkinsRepository()
}

func newCapesRepository(container *di.Container) (http.CapesRepository, error) {
	var factory dbModule.RepositoriesCreator
	err := container.Resolve(&factory, di.Name("fs"))
	if err != nil {
		return nil, err
	}

	return factory.CreateCapesRepository()
}

func newMojangUUIDsRepository(container *di.Container) (mojangtextures.UuidsStorage, error) {
	var factory dbModule.RepositoriesCreator
	err := container.Resolve(&factory, di.Name("redis"))
	if err != nil {
		return nil, err
	}

	return factory.CreateMojangUuidsRepository()
}

func newMojangSignedTexturesStorage() mojangtextures.TexturesStorage {
	texturesStorage := mojangtextures.NewInMemoryTexturesStorage()
	texturesStorage.Start()

	return texturesStorage
}
