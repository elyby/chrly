package ui

import (
	"github.com/mono83/slf/wd"
	"elyby/minecraft-skinsystem/repositories"
)

type uiService struct {
	logger wd.Watchdog
	skinsRepo repositories.SkinsRepository
	capesRepo repositories.CapesRepository
}

func NewUiService(
	logger wd.Watchdog,
	skinsRepo repositories.SkinsRepository,
	capesRepo repositories.CapesRepository,
) (*uiService, error) {
	return &uiService{
		logger: logger,
		skinsRepo: skinsRepo,
		capesRepo: capesRepo,
	}, nil
}
