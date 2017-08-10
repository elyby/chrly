package capes

import (
	"elyby/minecraft-skinsystem/repositories"
)

type CapesRepositoryCreator interface {
	Create() (repositories.CapesRepository, error)
}

type CapeNotFoundError struct {
	Who string
}

func (e CapeNotFoundError) Error() string {
	return "Cape file not found."
}
