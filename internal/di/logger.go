package di

import (
	"github.com/defval/di"
	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"

	"ely.by/chrly/internal/version"
)

var loggerDiOptions = di.Options(
	di.Provide(newSentry),
)

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
