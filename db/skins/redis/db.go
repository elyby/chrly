package redis

import (
	"elyby/minecraft-skinsystem/model"

	"github.com/mediocregopher/radix.v2/pool"
)

type Config struct {
	Addr string
	PollSize int
}

func (cfg *Config) CreateRepo() (model.SkinsRepository, error) {
	conn, err := pool.New("tcp", cfg.Addr, cfg.PollSize)
	if err != nil {
		return nil, err
	}

	// TODO: здесь можно запустить горутину по восстановлению соединения

	return &redisDb{conn: conn}, err
}
