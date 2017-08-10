package capes

import (
	"os"
	"path"
	"strings"

	"elyby/minecraft-skinsystem/model"
	"elyby/minecraft-skinsystem/repositories"
)

type FilesystemCapesFactory struct {
	StoragePath string
}

func (cfg *FilesystemCapesFactory) Create() (repositories.CapesRepository, error) {
	return &filesDb{path: cfg.StoragePath}, nil
}

type filesDb struct {
	path string
}

func (repository *filesDb) FindByUsername(username string) (model.Cape, error) {
	var record model.Cape
	if username == "" {
		return record, &CapeNotFoundError{username}
	}

	capePath := path.Join(repository.path, strings.ToLower(username) + ".png")
	file, err := os.Open(capePath)
	if err != nil {
		return record, &CapeNotFoundError{username}
	}

	record.File = file

	return record, nil
}
