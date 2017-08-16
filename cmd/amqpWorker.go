package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"elyby/minecraft-skinsystem/bootstrap"
	"elyby/minecraft-skinsystem/db"
	"elyby/minecraft-skinsystem/worker"
)

var amqpWorkerCmd = &cobra.Command{
	Use:   "amqp-worker",
	Short: "Launches a worker which listens to events and processes them",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := bootstrap.CreateLogger(viper.GetString("statsd.addr"))
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

		logger.Info("Initializing AMQP connection")
		amqpChannel, err := bootstrap.CreateRabbitMQChannel(&bootstrap.RabbitMQConfig{
			Host:     viper.GetString("amqp.host"),
			Port:     viper.GetInt("amqp.port"),
			Username: viper.GetString("amqp.username"),
			Password: viper.GetString("amqp.password"),
			Vhost:    viper.GetString("amqp.vhost"),
		})
		if err != nil {
			logger.Emergency(fmt.Sprintf("Error on connecting AMQP: %+v", err))
			return
		}
		logger.Info("AMQP connection successfully initialized")

		services := &worker.Services{
			Logger:    logger,
			Channel:   amqpChannel,
			SkinsRepo: skinsRepo,
		}

		if err := services.Run(); err != nil {
			logger.Error(fmt.Sprintf("Cannot initialize worker: %+v", err))
		}
	},
}

func init() {
	RootCmd.AddCommand(amqpWorkerCmd)
}
