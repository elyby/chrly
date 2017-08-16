package bootstrap

import (
	"fmt"
	"net/url"
	"os"

	"github.com/mono83/slf/rays"
	"github.com/mono83/slf/recievers/ansi"
	"github.com/mono83/slf/recievers/statsd"
	"github.com/mono83/slf/wd"
	"github.com/streadway/amqp"
)

func CreateLogger(statsdAddr string) (wd.Watchdog, error) {
	wd.AddReceiver(ansi.New(true, true, false))
	if statsdAddr != "" {
		hostname, _ := os.Hostname()
		statsdReceiver, err := statsd.NewReceiver(statsd.Config{
			Address: statsdAddr,
			Prefix: "ely.skinsystem." + hostname + ".app.",
			FlushEvery: 1,
		})

		if err != nil {
			return nil, err
		}

		wd.AddReceiver(statsdReceiver)
	}

	return wd.New("", "").WithParams(rays.Host), nil
}

type RabbitMQConfig struct {
	Username string
	Password string
	Host string
	Port int
	Vhost string
}

func CreateRabbitMQChannel(config *RabbitMQConfig) (*amqp.Channel, error) {
	addr := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		url.PathEscape(config.Vhost),
	)

	rabbitConnection, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	rabbitChannel, err := rabbitConnection.Channel()
	if err != nil {
		return nil, err
	}

	return rabbitChannel, nil
}
