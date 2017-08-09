package files

import (
	"elyby/minecraft-skinsystem/repositories"
)

type FilesystemCapesFactory struct {
	StoragePath string
}

func (cfg *FilesystemCapesFactory) Create() (repositories.CapesRepository, error) {
	return &filesDb{path: cfg.StoragePath}, nil
}
