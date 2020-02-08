package cmd

import (
	"fmt"
	"log"

	"github.com/mono83/slf/wd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/auth"
	"github.com/elyby/chrly/bootstrap"
	"github.com/elyby/chrly/db"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts HTTP handler for the skins system",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: this is a mess, need to organize this code somehow to make services initialization more compact
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

		uuidsProvider, err := bootstrap.CreateMojangUUIDsProvider(nil)
		if err != nil {
			logger.Emergency("Unable to parse remote url :err", wd.ErrParam(err))
			return
		}

		texturesStorage := mojangtextures.NewInMemoryTexturesStorage()
		texturesStorage.Start()
		mojangTexturesProvider := &mojangtextures.Provider{
			// TODO: configure emitter
			UUIDsProvider: uuidsProvider,
			TexturesProvider: &mojangtextures.MojangApiTexturesProvider{
				// TODO: configure emitter
			},
			Storage: &mojangtextures.SeparatedStorage{
				UuidsStorage:    mojangUuidsRepository,
				TexturesStorage: texturesStorage,
			},
		}
		logger.Info("Mojang's textures queue is successfully initialized")

		address := fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port"))
		handler := (&http.Skinsystem{
			SkinsRepo:               skinsRepo,
			CapesRepo:               capesRepo,
			MojangTexturesProvider:  mojangTexturesProvider,
			Auth:                    &auth.JwtAuth{Key: []byte(viper.GetString("chrly.secret"))},
			TexturesExtraParamName:  viper.GetString("textures.extra_param_name"),
			TexturesExtraParamValue: viper.GetString("textures.extra_param_value"),
		}).CreateHandler()

		finishChan := make(chan bool)
		go func() {
			logger.Info(fmt.Sprintf("Starting the app, HTTP on: %s", address))
			if err := http.Serve(address, handler); err != nil {
				logger.Error(fmt.Sprintf("Error in main(): %v", err))
				finishChan <- true
			}
		}()

		go func() {
			s := waitForExitSignal()
			logger.Info(fmt.Sprintf("Got signal: %v, exiting.", s))
			finishChan <- true
		}()

		<-finishChan
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
