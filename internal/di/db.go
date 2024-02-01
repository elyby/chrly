package di

import (
	"context"
	"fmt"

	"github.com/defval/di"
	"github.com/spf13/viper"

	db2 "github.com/elyby/chrly/internal/db"
	"github.com/elyby/chrly/internal/db/redis"
	es "github.com/elyby/chrly/internal/eventsubscribers"
	"github.com/elyby/chrly/internal/mojang"
	"github.com/elyby/chrly/internal/profiles"
)

// v4 had the idea that it would be possible to separate backends for storing skins and capes.
// But in v5 the storage will be unified, so this is just temporary constructors before large reworking.
//
// Since there are no options for selecting target backends,
// all constants in this case point to static specific implementations.
var db = di.Options(
	di.Provide(newRedis,
		di.As(new(profiles.ProfilesRepository)),
		di.As(new(profiles.ProfilesFinder)),
		di.As(new(mojang.MojangUuidsStorage)),
	),
)

func newRedis(container *di.Container, config *viper.Viper) (*redis.Redis, error) {
	config.SetDefault("storage.redis.host", "localhost")
	config.SetDefault("storage.redis.port", 6379)
	config.SetDefault("storage.redis.poolSize", 10)

	conn, err := redis.New(
		context.Background(),
		db2.NewZlibEncoder(&db2.JsonSerializer{}),
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
