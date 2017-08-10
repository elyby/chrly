package skins

import (
	"elyby/minecraft-skinsystem/repositories"
)

type SkinsRepositoryCreator interface {
	Create() (repositories.SkinsRepository, error)
}

type SkinNotFoundError struct {
	Who string
}

func (e SkinNotFoundError) Error() string {
	return "Skin data not found."
}
