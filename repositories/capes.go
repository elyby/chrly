package repositories

import "elyby/minecraft-skinsystem/model"

type CapesRepository interface {
	FindByUsername(username string) (model.Cape, error)
}
