package bootstrap

import (
	"github.com/elyby/chrly/http"
	"net/url"
	"os"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/mono83/slf/rays"
	"github.com/mono83/slf/recievers/sentry"
	"github.com/mono83/slf/recievers/statsd"
	"github.com/mono83/slf/recievers/writer"
	"github.com/mono83/slf/wd"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/mojangtextures"
	"github.com/elyby/chrly/version"
)

func CreateLogger(statsdAddr string, sentryAddr string) (wd.Watchdog, error) {
	wd.AddReceiver(writer.New(writer.Options{
		Marker:     false,
		TimeFormat: "15:04:05.000",
	}))

	if statsdAddr != "" {
		hostname, _ := os.Hostname()
		statsdReceiver, err := statsd.NewReceiver(statsd.Config{
			Address:    statsdAddr,
			Prefix:     "ely.skinsystem." + hostname + ".app.",
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
		programVersion := version.Version()
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

func init() {
	viper.SetDefault("queue.loop_delay", 2*time.Second+500*time.Millisecond)
	viper.SetDefault("queue.batch_size", 10)
}

func CreateMojangUUIDsProvider(emitter http.Emitter) (mojangtextures.UUIDsProvider, error) {
	var uuidsProvider mojangtextures.UUIDsProvider
	preferredUuidsProvider := viper.GetString("mojang_textures.uuids_provider.driver")
	if preferredUuidsProvider == "remote" {
		remoteUrl, err := url.Parse(viper.GetString("mojang_textures.uuids_provider.url"))
		if err != nil {
			return nil, err
		}

		uuidsProvider = &mojangtextures.RemoteApiUuidsProvider{
			Emitter: emitter,
			Url:     *remoteUrl,
		}
	} else {
		uuidsProvider = &mojangtextures.BatchUuidsProvider{
			Emitter:        emitter,
			IterationDelay: viper.GetDuration("queue.loop_delay"),
			IterationSize:  viper.GetInt("queue.batch_size"),
		}
	}

	return uuidsProvider, nil
}
