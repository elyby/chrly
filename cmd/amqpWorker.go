package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"elyby/minecraft-skinsystem/api/accounts"
	"elyby/minecraft-skinsystem/bootstrap"
	"elyby/minecraft-skinsystem/db"
	"elyby/minecraft-skinsystem/worker"
)

var amqpWorkerCmd = &cobra.Command{
	Use:   "amqp-worker",
	Short: "Launches a worker which listens to events and processes them",
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

		logger.Info("Creating AMQP client")
		amqpClient := bootstrap.CreateRabbitMQClient(&bootstrap.RabbitMQConfig{
			Host:     viper.GetString("amqp.host"),
			Port:     viper.GetInt("amqp.port"),
			Username: viper.GetString("amqp.username"),
			Password: viper.GetString("amqp.password"),
			Vhost:    viper.GetString("amqp.vhost"),
		})

		accountsApi := (&accounts.Config{
			Addr:   viper.GetString("api.accounts.host"),
			Id:     viper.GetString("api.accounts.id"),
			Secret: viper.GetString("api.accounts.secret"),
			Scopes: viper.GetStringSlice("api.accounts.scopes"),
		}).GetTokenWithAutoRefresh()

		services := &worker.Services{
			Logger:      logger,
			AmqpClient:  amqpClient,
			SkinsRepo:   skinsRepo,
			AccountsAPI: accountsApi,
		}

		if err := services.Run(); err != nil {
			logger.Error(fmt.Sprintf("Cannot initialize worker: %+v", err))
		}
	},
}

func init() {
	RootCmd.AddCommand(amqpWorkerCmd)
}
