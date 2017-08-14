package db

import (
	"github.com/spf13/viper"

	"elyby/minecraft-skinsystem/repositories"
)

type StorageFactory struct {
	Config *viper.Viper
}

type RepositoriesCreator interface {
	CreateSkinsRepository() (repositories.SkinsRepository, error)
	CreateCapesRepository() (repositories.CapesRepository, error)
}

func (factory *StorageFactory) CreateFactory(backend string) RepositoriesCreator {
	switch backend {
	case "redis":
		return &RedisFactory{
			Host:     factory.Config.GetString("storage.redis.host"),
			Port:     factory.Config.GetInt("storage.redis.port"),
			PoolSize: factory.Config.GetInt("storage.redis.poolSize"),
		}
	case "filesystem":
		return &FilesystemFactory{
			BasePath    : factory.Config.GetString("storage.filesystem.basePath"),
			CapesDirName: factory.Config.GetString("storage.filesystem.capesDirName"),
		}
	}

	return nil
}
