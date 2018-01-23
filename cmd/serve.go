package cmd

import (
	"fmt"
	"log"

	"elyby/minecraft-skinsystem/auth"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"elyby/minecraft-skinsystem/bootstrap"
	"elyby/minecraft-skinsystem/db"
	"elyby/minecraft-skinsystem/http"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Runs the system server skins",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := bootstrap.CreateLogger(viper.GetString("statsd.addr"), viper.GetString("sentry.dsn"))
		if err != nil {
			log.Fatal(fmt.Printf("Cannot initialize logger: %v", err))
		}
		logger.Info("Logger successfully initialized")

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

		cfg := &http.Config{
			ListenSpec: fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port")),
			SkinsRepo:  skinsRepo,
			CapesRepo:  capesRepo,
			Logger:     logger,
			Auth:       &auth.JwtAuth{},
		}

		if err := cfg.Run(); err != nil {
			logger.Error(fmt.Sprintf("Error in main(): %v", err))
		}
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
