package di

import "github.com/defval/di"

func New() (*di.Container, error) {
	container, err := di.New(
		config,
		dispatcher,
		logger,
		db,
		mojangTextures,
		handlers,
		profilesDi,
		server,
		signer,
	)
	if err != nil {
		return nil, err
	}

	return container, nil
}
