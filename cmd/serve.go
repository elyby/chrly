package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"

	"github.com/mono83/slf/rays"
	"github.com/mono83/slf/recievers/ansi"
	"github.com/mono83/slf/wd"
	"github.com/spf13/cobra"

	"elyby/minecraft-skinsystem/daemon"
	"elyby/minecraft-skinsystem/db/capes"
	"elyby/minecraft-skinsystem/db/skins"
	"elyby/minecraft-skinsystem/ui"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Запускает сервер системы скинов",
	Long: "Более длинное описание пока не было придумано",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: извлечь все инициализации зависимостей в парсер конфигурации

		// Logger
		wd.AddReceiver(ansi.New(true, true, false))
		logger := wd.New("", "").WithParams(rays.Host)

		// Skins repository
		logger.Info("Connecting to redis")
		skinsRepoCfg := &skins.RedisSkinsFactory{
			//Addr: "redis:6379",
			Addr: "localhost:16379",
			PollSize: 10,
		}
		skinsRepo, err := skinsRepoCfg.Create()
		if err != nil {
			logger.Emergency(fmt.Sprintf("Error on creating skins repo: %v", err))
			return
		}
		logger.Info("Successfully connected to redis")

		// Capes repository
		_, file, _, _ := runtime.Caller(0)
		capesRepoCfg := &capes.FilesystemCapesFactory{
			StoragePath: path.Join(filepath.Dir(file), "data/capes"),
		}
		capesRepo, err := capesRepoCfg.Create()
		if err != nil {
			logger.Emergency(fmt.Sprintf("Error on creating capes repo: %v", err))
			return
		}



		cfg := &daemon.Config{
			ListenSpec: "localhost:35644",
			SkinsRepo: skinsRepo,
			CapesRepo: capesRepo,
			Logger: logger,
			UI: ui.Config{},
		}

		if err := daemon.Run(cfg); err != nil {
			logger.Error(fmt.Sprintf("Error in main(): %v", err))
		}
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
