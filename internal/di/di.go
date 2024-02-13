package di

import "github.com/defval/di"

func New() (*di.Container, error) {
	return di.New(
		configDiOptions,
		contextDiOptions,
		dbDiOptions,
		handlersDiOptions,
		loggerDiOptions,
		mojangDiOptions,
		profilesDiOptions,
		securityDiOptions,
		serverDiOptions,
	)
}
