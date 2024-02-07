package di

import (
	"context"
	"fmt"

	"github.com/defval/di"
	"github.com/etherlabsio/healthcheck/v2"
	"github.com/spf13/viper"

	"ely.by/chrly/internal/db"
	"ely.by/chrly/internal/db/redis"
	"ely.by/chrly/internal/mojang"
	"ely.by/chrly/internal/profiles"
)

// Since there are no options for selecting target backends,
// all constants in this case point to static specific implementations.
var dbDeOptions = di.Options(
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
		db.NewZlibEncoder(&db.JsonSerializer{}),
		fmt.Sprintf("%s:%d", config.GetString("storage.redis.host"), config.GetInt("storage.redis.port")),
		config.GetInt("storage.redis.poolSize"),
	)
	if err != nil {
		return nil, err
	}

	if err := container.Provide(func() *namedHealthChecker {
		return &namedHealthChecker{
			Name:    "redis",
			Checker: healthcheck.CheckerFunc(conn.Ping),
		}
	}); err != nil {
		return nil, err
	}

	return conn, nil
}
