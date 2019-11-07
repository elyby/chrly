package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/api/mojang/queue"
	"github.com/elyby/chrly/auth"
	"github.com/elyby/chrly/bootstrap"
	"github.com/elyby/chrly/db"
	"github.com/elyby/chrly/http"
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
		redisFactory := storageFactory.CreateFactory("redis")
		skinsRepo, err := redisFactory.CreateSkinsRepository()
		if err != nil {
			logger.Emergency(fmt.Sprintf("Error on creating skins repo: %+v", err))
			return
		}
		logger.Info("Skins repository successfully initialized")

		logger.Info("Initializing capes repository")
		filesystemFactory := storageFactory.CreateFactory("filesystem")
		capesRepo, err := filesystemFactory.CreateCapesRepository()
		if err != nil {
			logger.Emergency(fmt.Sprintf("Error on creating capes repo: %v", err))
			return
		}
		logger.Info("Capes repository successfully initialized")

		logger.Info("Preparing Mojang's textures queue")
		mojangUuidsRepository, err := redisFactory.CreateMojangUuidsRepository()
		if err != nil {
			logger.Emergency(fmt.Sprintf("Error on creating mojang uuids repo: %v", err))
			return
		}

		queue.UuidsQueueIterationDelay = time.Duration(viper.GetInt("queue.loop_delay")) * time.Millisecond
		texturesStorage := queue.CreateInMemoryTexturesStorage()
		texturesStorage.Start()
		mojangTexturesQueue := &queue.JobsQueue{
			Logger: logger,
			Storage: &queue.SplittedStorage{
				UuidsStorage:    mojangUuidsRepository,
				TexturesStorage: texturesStorage,
			},
		}
		logger.Info("Mojang's textures queue is successfully initialized")

		cfg := &http.Config{
			ListenSpec:          fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port")),
			SkinsRepo:           skinsRepo,
			CapesRepo:           capesRepo,
			MojangTexturesQueue: mojangTexturesQueue,
			Logger:              logger,
			Auth:                &auth.JwtAuth{Key: []byte(viper.GetString("chrly.secret"))},
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
	viper.SetDefault("queue.loop_delay", 2_500)
}
