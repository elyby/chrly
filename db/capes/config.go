package capes

import "elyby/minecraft-skinsystem/model"

type CapesRepositoryConfig interface {
	CreateRepo() (model.CapesRepository, error)
}
