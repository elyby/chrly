package di

import (
	"github.com/defval/di"
	"github.com/getsentry/raven-go"
	"github.com/mono83/slf"
	"github.com/mono83/slf/rays"
	"github.com/mono83/slf/recievers/sentry"
	"github.com/mono83/slf/recievers/writer"
	"github.com/mono83/slf/wd"
	"github.com/spf13/viper"

	"ely.by/chrly/internal/version"
)

var loggerDiOptions = di.Options(
	di.Provide(newLogger),
	di.Provide(newSentry),
)

type loggerParams struct {
	di.Inject

	SentryRaven *raven.Client `di:"" optional:"true"`
}

func newLogger(params loggerParams) slf.Logger {
	dispatcher := &slf.Dispatcher{}
	dispatcher.AddReceiver(writer.New(writer.Options{
		Marker:     false,
		TimeFormat: "15:04:05.000",
	}))

	if params.SentryRaven != nil {
		sentryReceiver, _ := sentry.NewReceiverWithCustomRaven(
			params.SentryRaven,
			&sentry.Config{
				MinLevel: "warn",
			},
		)
		dispatcher.AddReceiver(sentryReceiver)
	}

	logger := wd.Custom("", "", dispatcher)
	logger.WithParams(rays.Host)

	return logger
}

func newSentry(config *viper.Viper) (*raven.Client, error) {
	sentryAddr := config.GetString("sentry.dsn")
	if sentryAddr == "" {
		return nil, nil
	}

	ravenClient, err := raven.New(sentryAddr)
	if err != nil {
		return nil, err
	}

	ravenClient.SetEnvironment("production")
	ravenClient.SetDefaultLoggerName("sentry-watchdog-receiver")
	ravenClient.SetRelease(version.Version())

	raven.DefaultClient = ravenClient

	return ravenClient, nil
}
