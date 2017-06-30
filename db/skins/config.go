package skins

import "elyby/minecraft-skinsystem/model"

type SkinsRepositoryConfig interface {
	CreateRepo() (model.SkinsRepository, error)
}
