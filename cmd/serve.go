package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/mono83/slf/wd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/bootstrap"
	"github.com/elyby/chrly/db"
	"github.com/elyby/chrly/eventsubscribers"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts HTTP handler for the skins system",
	Run: func(cmd *cobra.Command, args []string) {
		dispatcher := bootstrap.CreateEventDispatcher()

		// TODO: this is a mess, need to organize this code somehow to make services initialization more compact
		logger, err := bootstrap.CreateLogger(viper.GetString("sentry.dsn"))
		if err != nil {
			log.Fatalf("Cannot initialize logger: %v", err)
		}
		logger.Info("Logger successfully initialized")

		(&eventsubscribers.Logger{Logger: logger}).ConfigureWithDispatcher(dispatcher)

		statsdAddr := viper.GetString("statsd.addr")
		if statsdAddr != "" {
			statsdReporter, err := bootstrap.CreateStatsReceiver(statsdAddr)
			if err != nil {
				logger.Emergency("Invalid statsd configuration :err", wd.ErrParam(err))
				os.Exit(1)
			}

			(&eventsubscribers.StatsReporter{StatsReporter: statsdReporter}).ConfigureWithDispatcher(dispatcher)
		}

		storageFactory := db.StorageFactory{Config: viper.GetViper()}

		logger.Info("Initializing skins repository")
		redisFactory := storageFactory.CreateFactory("redis")
		skinsRepo, err := redisFactory.CreateSkinsRepository()
		if err != nil {
			logger.Emergency("Error on creating skins repo: :err", wd.ErrParam(err))
			os.Exit(1)
		}
		logger.Info("Skins repository successfully initialized")

		logger.Info("Initializing capes repository")
		filesystemFactory := storageFactory.CreateFactory("filesystem")
		capesRepo, err := filesystemFactory.CreateCapesRepository()
		if err != nil {
			logger.Emergency("Error on creating capes repo: :err", wd.ErrParam(err))
			os.Exit(1)
		}
		logger.Info("Capes repository successfully initialized")

		logger.Info("Preparing Mojang's textures queue")
		mojangUuidsRepository, err := redisFactory.CreateMojangUuidsRepository()
		if err != nil {
			logger.Emergency("Error on creating mojang uuids repo: :err", wd.ErrParam(err))
			os.Exit(1)
		}

		uuidsProvider, err := bootstrap.CreateMojangUUIDsProvider(dispatcher)
		if err != nil {
			logger.Emergency("Unable to parse remote url :err", wd.ErrParam(err))
			os.Exit(1)
		}

		texturesStorage := mojangtextures.NewInMemoryTexturesStorage()
		texturesStorage.Start()
		mojangTexturesProvider := &mojangtextures.Provider{
			Emitter:       dispatcher,
			UUIDsProvider: uuidsProvider,
			TexturesProvider: &mojangtextures.MojangApiTexturesProvider{
				Emitter: dispatcher,
			},
			Storage: &mojangtextures.SeparatedStorage{
				UuidsStorage:    mojangUuidsRepository,
				TexturesStorage: texturesStorage,
			},
		}
		logger.Info("Mojang's textures queue is successfully initialized")

		address := fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port"))
		handler := (&http.Skinsystem{
			Emitter:                 dispatcher,
			SkinsRepo:               skinsRepo,
			CapesRepo:               capesRepo,
			MojangTexturesProvider:  mojangTexturesProvider,
			Authenticator:           &http.JwtAuth{Key: []byte(viper.GetString("chrly.secret"))},
			TexturesExtraParamName:  viper.GetString("textures.extra_param_name"),
			TexturesExtraParamValue: viper.GetString("textures.extra_param_value"),
		}).CreateHandler()

		finishChan := make(chan bool)
		go func() {
			logger.Info("Starting the app, HTTP on: :addr", wd.StringParam("addr", address))
			if err := http.Serve(address, handler); err != nil {
				logger.Emergency("Error in main(): :err", wd.ErrParam(err))
				finishChan <- true
			}
		}()

		go func() {
			s := waitForExitSignal()
			logger.Info("Got signal: :signal, exiting", wd.StringParam("signal", s.String()))
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
