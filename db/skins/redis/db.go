package redis

import (
	"elyby/minecraft-skinsystem/repositories"
	"github.com/mediocregopher/radix.v2/pool"
)

type RedisSkinsFactory struct {
	Addr string
	PollSize int
}

func (cfg *RedisSkinsFactory) Create() (repositories.SkinsRepository, error) {
	conn, err := pool.New("tcp", cfg.Addr, cfg.PollSize)
	if err != nil {
		return nil, err
	}

	// TODO: здесь можно запустить горутину по восстановлению соединения

	return &redisDb{conn: conn}, nil
}
