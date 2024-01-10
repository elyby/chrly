package di

import (
	"context"
	"fmt"
	"path"

	"github.com/defval/di"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/db/fs"
	"github.com/elyby/chrly/db/redis"
	es "github.com/elyby/chrly/eventsubscribers"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojang"
)

// v4 had the idea that it would be possible to separate backends for storing skins and capes.
// But in v5 the storage will be unified, so this is just temporary constructors before large reworking.
//
// Since there are no options for selecting target backends,
// all constants in this case point to static specific implementations.
var db = di.Options(
	di.Provide(newRedis,
		di.As(new(http.SkinsRepository)),
		di.As(new(mojang.MojangUuidsStorage)),
	),
	di.Provide(newFSFactory,
		di.As(new(http.CapesRepository)),
	),
)

func newRedis(container *di.Container, config *viper.Viper) (*redis.Redis, error) {
	config.SetDefault("storage.redis.host", "localhost")
	config.SetDefault("storage.redis.port", 6379)
	config.SetDefault("storage.redis.poolSize", 10)

	conn, err := redis.New(
		context.Background(),
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
