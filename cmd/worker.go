package cmd

import (
	"fmt"
	"log"

	"github.com/mono83/slf/wd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/bootstrap"
	"github.com/elyby/chrly/http"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Starts HTTP handler for the Mojang usernames to UUIDs worker",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := bootstrap.CreateLogger(viper.GetString("statsd.addr"), viper.GetString("sentry.dsn"))
		if err != nil {
			log.Fatal(fmt.Printf("Cannot initialize logger: %v", err))
		}
		logger.Info("Logger successfully initialized")

		uuidsProvider, err := bootstrap.CreateMojangUUIDsProvider(logger)
		if err != nil {
			logger.Emergency("Unable to parse remote url :err", wd.ErrParam(err))
			return
		}

		cfg := &http.UUIDsWorker{
			ListenSpec:    fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port")),
			UUIDsProvider: uuidsProvider,
			Logger:        logger,
		}

		finishChan := make(chan bool)
		go func() {
			if err := cfg.Run(); err != nil {
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
	RootCmd.AddCommand(workerCmd)
}
