package di

import "github.com/defval/di"

func New() (*di.Container, error) {
	return di.New(
		configDiOptions,
		loggerDiOptions,
		dbDeOptions,
		mojangDiOptions,
		handlersDiOptions,
		profilesDiOptions,
		serverDiOptions,
		securityDiOptions,
	)
}
