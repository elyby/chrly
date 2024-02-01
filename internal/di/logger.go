package di

import (
	"os"

	"github.com/defval/di"
	"github.com/getsentry/raven-go"
	"github.com/mono83/slf"
	"github.com/mono83/slf/rays"
	"github.com/mono83/slf/recievers/sentry"
	"github.com/mono83/slf/recievers/statsd"
	"github.com/mono83/slf/recievers/writer"
	"github.com/mono83/slf/wd"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/internal/eventsubscribers"
	"github.com/elyby/chrly/internal/version"
)

var logger = di.Options(
	di.Provide(newLogger),
	di.Provide(newSentry),
	di.Provide(newStatsReporter),
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

func newStatsReporter(config *viper.Viper) (slf.StatsReporter, error) {
	dispatcher := &slf.Dispatcher{}

	statsdAddr := config.GetString("statsd.addr")
	if statsdAddr != "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		statsdReceiver, err := statsd.NewReceiver(statsd.Config{
			Address:    statsdAddr,
			Prefix:     "ely.skinsystem." + hostname + ".app.",
			FlushEvery: 1,
		})
		if err != nil {
			return nil, err
		}

		dispatcher.AddReceiver(statsdReceiver)
	}

	return wd.Custom("", "", dispatcher), nil
}

func enableReporters(reporter slf.StatsReporter, factories []eventsubscribers.Reporter) {
	for _, factory := range factories {
		factory.Enable(reporter)
	}
}
