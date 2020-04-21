package di

import (
	"fmt"
	"path"

	"github.com/goava/di"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/db/fs"
	"github.com/elyby/chrly/db/redis"
	es "github.com/elyby/chrly/eventsubscribers"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

// v4 had the idea that it would be possible to separate backends for storing skins and capes.
// But in v5 the storage will be unified, so this is just temporary constructors before large reworking.
//
// Since there are no options for selecting target backends,
// all constants in this case point to static specific implementations.
var db = di.Options(
	di.Provide(newRedis,
		di.As(new(http.SkinsRepository)),
		di.As(new(mojangtextures.UuidsStorage)),
	),
	di.Provide(newFSFactory,
		di.As(new(http.CapesRepository)),
	),
	di.Provide(newMojangSignedTexturesStorage),
)

func newRedis(container *di.Container, config *viper.Viper) (*redis.Redis, error) {
	config.SetDefault("storage.redis.host", "localhost")
	config.SetDefault("storage.redis.port", 6379)
	config.SetDefault("storage.redis.poll", 10)

	conn, err := redis.New(
		fmt.Sprintf("%s:%d", config.GetString("storage.redis.host"), config.GetInt("storage.redis.port")),
		config.GetInt("storage.redis.poolSize"),
	)
	if err != nil {
		return nil, err
	}

	if err := container.Provide(func() *namedHealthChecker {
		return &namedHealthChecker{
			Name:    "redis",
			Checker: es.DatabaseChecker(conn),
		}
	}); err != nil {
		return nil, err
	}

	return conn, nil
}

func newFSFactory(config *viper.Viper) (*fs.Filesystem, error) {
	config.SetDefault("storage.filesystem.basePath", "data")
	config.SetDefault("storage.filesystem.capesDirName", "capes")

	return fs.New(path.Join(
		config.GetString("storage.filesystem.basePath"),
		config.GetString("storage.filesystem.capesDirName"),
	))
}

func newMojangSignedTexturesStorage() mojangtextures.TexturesStorage {
	texturesStorage := mojangtextures.NewInMemoryTexturesStorage()
	texturesStorage.Start()

	return texturesStorage
}
