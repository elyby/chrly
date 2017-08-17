package ui

import (
	"github.com/mono83/slf/wd"
	"elyby/minecraft-skinsystem/interfaces"
)

type uiService struct {
	logger    wd.Watchdog
	skinsRepo interfaces.SkinsRepository
	capesRepo interfaces.CapesRepository
}

func NewUiService(
	logger wd.Watchdog,
	skinsRepo interfaces.SkinsRepository,
	capesRepo interfaces.CapesRepository,
) (*uiService, error) {
	return &uiService{
		logger: logger,
		skinsRepo: skinsRepo,
		capesRepo: capesRepo,
	}, nil
}
