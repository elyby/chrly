package db

import (
	"os"
	"path"
	"strings"

	"github.com/elyby/chrly/interfaces"
	"github.com/elyby/chrly/model"
)

type FilesystemFactory struct {
	BasePath 	 string
	CapesDirName string
}

func (f FilesystemFactory) CreateSkinsRepository() (interfaces.SkinsRepository, error) {
	panic("skins repository not supported for this storage type")
}

func (f FilesystemFactory) CreateCapesRepository() (interfaces.CapesRepository, error) {
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

func (repository *filesStorage) FindByUsername(username string) (*model.Cape, error) {
	if username == "" {
		return nil, &CapeNotFoundError{username}
	}

	capePath := path.Join(repository.path, strings.ToLower(username) + ".png")
	file, err := os.Open(capePath)
	if err != nil {
		return nil, &CapeNotFoundError{username}
	}

	return &model.Cape{
		File: file,
	}, nil
}
