package di

import "github.com/goava/di"

func New() (*di.Container, error) {
	container, err := di.New(
		di.WithCompile(),
		config,
		dispatcher,
		logger,
		db,
		mojangTextures,
	)
	if err != nil {
		return nil, err
	}

	// Inject container itself into dependencies graph
	// See https://github.com/goava/di/issues/8#issuecomment-614227320
	err = container.Provide(func() *di.Container {
		return container
	})
	if err != nil {
		return nil, err
	}

	err = container.Compile()
	if err != nil {
		return nil, err
	}

	return container, nil
}
