package di

import (
	"github.com/goava/di"
	"github.com/spf13/viper"

	. "github.com/elyby/chrly/db"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

var db = di.Options(
	di.Provide(newRedisFactory),
	di.Provide(newFSFactory),
	di.Provide(newSkinsRepository),
	di.Provide(newCapesRepository),
	di.Provide(newMojangUUIDsRepository),
	di.Provide(newMojangSignedTexturesStorage),
)

func newRedisFactory(config *viper.Viper) *RedisFactory {
	config.SetDefault("storage.redis.host", "localhost")
	config.SetDefault("storage.redis.port", 6379)
	config.SetDefault("storage.redis.poll", 10)

	return &RedisFactory{
		Host:     config.GetString("storage.redis.host"),
		Port:     config.GetInt("storage.redis.port"),
		PoolSize: config.GetInt("storage.redis.poolSize"),
	}
}

func newFSFactory(config *viper.Viper) *FilesystemFactory {
	config.SetDefault("storage.filesystem.basePath", "data")
	config.SetDefault("storage.filesystem.capesDirName", "capes")

	return &FilesystemFactory{
		BasePath:     config.GetString("storage.filesystem.basePath"),
		CapesDirName: config.GetString("storage.filesystem.capesDirName"),
	}
}

// v4 had the idea that it would be possible to separate backends for storing skins and capes.
// But in v5 the storage will be unified, so this is just temporary constructors before large reworking.
//
// Since there are no options for selecting target backends,
// all constants in this case point to static specific implementations.

func newSkinsRepository(factory *RedisFactory) (http.SkinsRepository, error) {
	return factory.CreateSkinsRepository()
}

func newCapesRepository(factory *FilesystemFactory) (http.CapesRepository, error) {
	return factory.CreateCapesRepository()
}

func newMojangUUIDsRepository(factory *RedisFactory) (mojangtextures.UuidsStorage, error) {
	return factory.CreateMojangUuidsRepository()
}

func newMojangSignedTexturesStorage() mojangtextures.TexturesStorage {
	texturesStorage := mojangtextures.NewInMemoryTexturesStorage()
	texturesStorage.Start()

	return texturesStorage
}
