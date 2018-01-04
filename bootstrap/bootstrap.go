package bootstrap

import (
	"fmt"
	"net/url"
	"os"

	"github.com/assembla/cony"
	"github.com/getsentry/raven-go"
	"github.com/mono83/slf/rays"
	"github.com/mono83/slf/recievers/sentry"
	"github.com/mono83/slf/recievers/statsd"
	"github.com/mono83/slf/recievers/writer"
	"github.com/mono83/slf/wd"
)

var version = ""

func GetVersion() string {
	return version
}

func CreateLogger(statsdAddr string, sentryAddr string) (wd.Watchdog, error) {
	wd.AddReceiver(writer.New(writer.Options{
		Marker: false,
		TimeFormat: "15:04:05.000",
	}))
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

	if sentryAddr != "" {
		ravenClient, err := raven.New(sentryAddr)
		if err != nil {
			return nil, err
		}

		ravenClient.SetEnvironment("production")
		ravenClient.SetDefaultLoggerName("sentry-watchdog-receiver")
		programVersion := GetVersion()
		if programVersion != "" {
			raven.SetRelease(programVersion)
		}

		sentryReceiver, err := sentry.NewReceiverWithCustomRaven(ravenClient, &sentry.Config{
			MinLevel: "warn",
		})
		if err != nil {
			return nil, err
		}

		wd.AddReceiver(sentryReceiver)
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

func CreateRabbitMQClient(config *RabbitMQConfig) *cony.Client {
	addr := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		url.PathEscape(config.Vhost),
	)

	client := cony.NewClient(cony.URL(addr), cony.Backoff(cony.DefaultBackoff))

	return client
}
