package files

import "elyby/minecraft-skinsystem/model"

type FilesystemCapesFactory struct {
	StoragePath string
}

func (cfg *FilesystemCapesFactory) Create() (model.CapesRepository, error) {
	return &filesDb{path: cfg.StoragePath}, nil
}
