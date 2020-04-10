package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/etherlabsio/healthcheck"
	"github.com/mono83/slf/wd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/bootstrap"
	"github.com/elyby/chrly/eventsubscribers"
	"github.com/elyby/chrly/http"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Starts HTTP handler for the Mojang usernames to UUIDs worker",
	Run: func(cmd *cobra.Command, args []string) {
		dispatcher := bootstrap.CreateEventDispatcher()

		// TODO: need to find a way to unify this initialization with the serve command
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

		uuidsProvider, err := bootstrap.CreateMojangUUIDsProvider(dispatcher)
		if err != nil {
			logger.Emergency("Unable to parse remote url :err", wd.ErrParam(err))
			os.Exit(1)
		}

		address := fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port"))
		handler := (&http.UUIDsWorker{
			Emitter:       dispatcher,
			UUIDsProvider: uuidsProvider,
		}).CreateHandler()
		handler.Handle("/healthcheck", healthcheck.Handler(
			healthcheck.WithChecker(
				"mojang-batch-uuids-provider-response",
				eventsubscribers.MojangBatchUuidsProviderChecker(
					dispatcher,
					viper.GetDuration("healthcheck.mojang_batch_uuids_provider_cool_down"),
				),
			),
		)).Methods("GET")

		finishChan := make(chan bool)
		go func() {
			logger.Info("Starting the worker, HTTP on: :addr", wd.StringParam("addr", address))
			if err := http.Serve(address, handler); err != nil {
				logger.Error("Error in main(): :err", wd.ErrParam(err))
				finishChan <- true
			}
		}()

		go func() {
			s := waitForExitSignal()
			logger.Info("Got signal: :code, exiting.", wd.StringParam("code", s.String()))
			finishChan <- true
		}()

		<-finishChan
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)
	viper.SetDefault("healthcheck.mojang_batch_uuids_provider_cool_down", time.Minute)
}
