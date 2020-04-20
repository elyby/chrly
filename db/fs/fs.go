package fs

import (
	"os"
	"path"
	"strings"

	"github.com/elyby/chrly/model"
)

func New(basePath string) (*Filesystem, error) {
	return &Filesystem{path: basePath}, nil
}

type Filesystem struct {
	path string
}

func (f *Filesystem) FindCapeByUsername(username string) (*model.Cape, error) {
	if username == "" {
		return nil, nil
	}

	capePath := path.Join(f.path, strings.ToLower(username)+".png")
	file, err := os.Open(capePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	return &model.Cape{
		File: file,
	}, nil
}
