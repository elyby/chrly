package files

import (
	"os"
	"path"
	"strings"

	"elyby/minecraft-skinsystem/model"
)

type filesDb struct {
	path string
}

func (repository *filesDb) FindByUsername(username string) (model.Cape, error) {
	var record model.Cape
	capePath := path.Join(repository.path, strings.ToLower(username) + ".png")
	file, err := os.Open(capePath)
	if err != nil {
		return record, CapeNotFound{username}
	}

	record.File = file

	return record, nil
}
