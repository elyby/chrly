package capes

import (
	"elyby/minecraft-skinsystem/repositories"
)

type CapesRepositoryCreator interface {
	Create() (repositories.CapesRepository, error)
}
