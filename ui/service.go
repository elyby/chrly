package ui

import (
	"elyby/minecraft-skinsystem/model"

	"github.com/mono83/slf/wd"
)

type uiService struct {
	logger wd.Watchdog
	skinsRepo model.SkinsRepository
	capesRepo model.CapesRepository
}

func NewUiService(
	logger wd.Watchdog,
	skinsRepo model.SkinsRepository,
	capesRepo model.CapesRepository,
) (*uiService, error) {
	return &uiService{
		logger: logger,
		skinsRepo: skinsRepo,
		capesRepo: capesRepo,
	}, nil
}
