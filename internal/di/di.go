package di

import "github.com/defval/di"

func New() (*di.Container, error) {
	return di.New(
		config,
		dispatcher,
		logger,
		db,
		mojangTextures,
		handlers,
		profilesDi,
		server,
		securityDiOptions,
	)
}
