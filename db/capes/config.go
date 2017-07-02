package capes

import "elyby/minecraft-skinsystem/model"

type CapesRepositoryCreator interface {
	Create() (model.CapesRepository, error)
}
