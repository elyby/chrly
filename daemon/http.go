package daemon

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mono83/slf/wd"

	"elyby/minecraft-skinsystem/repositories"
	"elyby/minecraft-skinsystem/ui"
)

type Config struct {
	ListenSpec string

	SkinsRepo  repositories.SkinsRepository
	CapesRepo  repositories.CapesRepository
	Logger     wd.Watchdog
	UI         ui.Config
}

func Run(cfg *Config) error {
	cfg.Logger.Info(fmt.Sprintf("Starting, HTTP on: %s\n", cfg.ListenSpec))

	uiService, err := ui.NewUiService(cfg.Logger, cfg.SkinsRepo, cfg.CapesRepo)
	if err != nil {
		cfg.Logger.Error(fmt.Sprintf("Error creating ui services: %v\n", err))
		return err
	}

	listener, err := net.Listen("tcp", cfg.ListenSpec)
	if err != nil {
		cfg.Logger.Error(fmt.Sprintf("Error creating listener: %v\n", err))
		return err
	}

	ui.Start(cfg.UI, uiService, listener)

	waitForSignal(cfg)

	return nil
}

func waitForSignal(cfg *Config) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	s := <-ch
	cfg.Logger.Info(fmt.Sprintf("Got signal: %v, exiting.", s))
}
