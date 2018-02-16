package interfaces

import (
	"github.com/elyby/chrly/model"
)

type SkinsRepository interface {
	FindByUsername(username string) (*model.Skin, error)
	FindByUserId(id int) (*model.Skin, error)
	Save(skin *model.Skin) error
	RemoveByUserId(id int) error
	RemoveByUsername(username string) error
}

type CapesRepository interface {
	FindByUsername(username string) (*model.Cape, error)
}
