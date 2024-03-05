package cmd

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ely.by/chrly/internal/di"
	"ely.by/chrly/internal/http"
	"ely.by/chrly/internal/otel"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts HTTP handler for the skins system",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startServer(di.ModuleSkinsystem, di.ModuleProfiles, di.ModuleSigner)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

func startServer(modules ...string) error {
	container := shouldGetContainer()

	var globalCtx context.Context
	err := container.Resolve(&globalCtx)
	if err != nil {
		return err
	}

	var config *viper.Viper
	err = container.Resolve(&config)
	if err != nil {
		return err
	}

	if !config.GetBool("otel.sdk.disabled") {
		shutdownOtel, err := otel.SetupOTelSDK(globalCtx)
		defer func() {
			err := shutdownOtel(context.Background())
			if err != nil {
				slog.Error("Unable to shutdown OpenTelemetry", slog.Any("error", err))
			}
		}()
		if err != nil {
			return err
		}
	}

	config.Set("modules", modules)

	err = container.Invoke(http.StartServer)
	if err != nil {
		return err
	}

	return nil
}
