package db

import (
	"os"
	"path"
	"strings"

	"elyby/minecraft-skinsystem/model"
	"elyby/minecraft-skinsystem/repositories"
)

type FilesystemFactory struct {
	BasePath 	 string
	CapesDirName string
}

func (f FilesystemFactory) CreateSkinsRepository() (repositories.SkinsRepository, error) {
	panic("skins repository not supported for this storage type")
}

func (f FilesystemFactory) CreateCapesRepository() (repositories.CapesRepository, error) {
	if err := f.validateFactoryConfig(); err != nil {
		return nil, err
	}

	return &filesStorage{path: path.Join(f.BasePath, f.CapesDirName)}, nil
}

func (f FilesystemFactory) validateFactoryConfig() error {
	if f.BasePath == "" {
		return &ParamRequired{"basePath"}
	}

	if f.CapesDirName == "" {
		f.CapesDirName = "capes"
	}

	return nil
}

type filesStorage struct {
	path string
}

func (repository *filesStorage) FindByUsername(username string) (model.Cape, error) {
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
