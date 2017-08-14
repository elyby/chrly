package cmd

import (
	"fmt"

	"github.com/mono83/slf/rays"
	"github.com/mono83/slf/recievers/ansi"
	"github.com/mono83/slf/wd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"elyby/minecraft-skinsystem/daemon"
	"elyby/minecraft-skinsystem/db"
	"elyby/minecraft-skinsystem/ui"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Запускает сервер системы скинов",
	Long: "Более длинное описание пока не было придумано",
	Run: func(cmd *cobra.Command, args []string) {
		wd.AddReceiver(ansi.New(true, true, false))
		logger := wd.New("", "").WithParams(rays.Host)

		storageFactory := db.StorageFactory{Config: viper.GetViper()}

		logger.Info("Initializing skins repository")
		skinsRepo, err := storageFactory.CreateFactory("redis").CreateSkinsRepository()
		if err != nil {
			logger.Emergency(fmt.Sprintf("Error on creating skins repo: %+v", err))
			return
		}
		logger.Info("Skins repository successfully initialized")

		logger.Info("Initializing capes repository")
		capesRepo, err := storageFactory.CreateFactory("filesystem").CreateCapesRepository()
		if err != nil {
			logger.Emergency(fmt.Sprintf("Error on creating capes repo: %v", err))
			return
		}
		logger.Info("Capes repository successfully initialized")

		cfg := &daemon.Config{
			ListenSpec: fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port")),
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
