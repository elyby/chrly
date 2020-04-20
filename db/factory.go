package db

import (
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

type RepositoriesCreator interface {
	CreateSkinsRepository() (http.SkinsRepository, error)
	CreateCapesRepository() (http.CapesRepository, error)
	CreateMojangUuidsRepository() (mojangtextures.UuidsStorage, error)
}
