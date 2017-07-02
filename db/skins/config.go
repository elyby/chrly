package skins

import "elyby/minecraft-skinsystem/model"

type SkinsRepositoryCreator interface {
	Create() (model.SkinsRepository, error)
}
