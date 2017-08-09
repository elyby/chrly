package repositories

import "elyby/minecraft-skinsystem/model"

type SkinsRepository interface {
	FindByUsername(username string) (model.Skin, error)
	FindByUserId(id int) (model.Skin, error)
}
