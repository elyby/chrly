package skins

import (
	"elyby/minecraft-skinsystem/repositories"
)

type SkinsRepositoryCreator interface {
	Create() (repositories.SkinsRepository, error)
}
