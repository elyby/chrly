package di

import "github.com/goava/di"

func New() (*di.Container, error) {
	container, err := di.New(
		config,
		dispatcher,
		logger,
		db,
		mojangTextures,
		handlers,
		server,
	)
	if err != nil {
		return nil, err
	}

	return container, nil
}
