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
	Short: "Starts http handler for the skins system",
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
			Auth:       &auth.JwtAuth{Key: []byte(viper.GetString("chrly.secret"))},
		}

		if err := cfg.Run(); err != nil {
			logger.Error(fmt.Sprintf("Error in main(): %v", err))
		}
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
	viper.SetDefault("server.host", "")
	viper.SetDefault("server.port", 80)
	viper.SetDefault("storage.redis.host", "localhost")
	viper.SetDefault("storage.redis.port", 6379)
	viper.SetDefault("storage.redis.poll", 10)
	viper.SetDefault("storage.filesystem.basePath", "data")
	viper.SetDefault("storage.filesystem.capesDirName", "capes")
}
