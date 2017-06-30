package files

import "elyby/minecraft-skinsystem/model"

type Config struct {
	StoragePath string
}

func (cfg *Config) CreateRepo() (model.CapesRepository, error) {
	return &filesDb{path: cfg.StoragePath}, nil
}
