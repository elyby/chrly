package di

import (
	"net/http"
	"strings"

	"github.com/etherlabsio/healthcheck"
	"github.com/goava/di"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	. "github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

var handlers = di.Options(
	di.Provide(newHandlerFactory, di.As(new(http.Handler))),
	di.Provide(newSkinsystemHandler, di.WithName("skinsystem")),
	di.Provide(newApiHandler, di.WithName("api")),
	di.Provide(newUUIDsWorkerHandler, di.WithName("worker")),
)

func newHandlerFactory(
	container *di.Container,
	config *viper.Viper,
	emitter Emitter,
) (*mux.Router, error) {
	enabledModules := config.GetStringSlice("modules")

	// gorilla.mux has no native way to combine multiple routers.
	// The hack used later in the code works for prefixes in addresses, but leads to misbehavior
	// if you set an empty prefix. Since the main application should be mounted at the root prefix,
	// we use it as the base router
	var router *mux.Router
	if hasValue(enabledModules, "skinsystem") {
		if err := container.Resolve(&router, di.Name("skinsystem")); err != nil {
			return nil, err
		}
	} else {
		router = mux.NewRouter()
	}

	router.StrictSlash(true)
	requestEventsMiddleware := CreateRequestEventsMiddleware(emitter, "skinsystem")
	router.Use(requestEventsMiddleware)
	// NotFoundHandler doesn't call for registered middlewares, so we must wrap it manually.
	// See https://github.com/gorilla/mux/issues/416#issuecomment-600079279
	router.NotFoundHandler = requestEventsMiddleware(http.HandlerFunc(NotFoundHandler))

	// Enable the worker module before api to allow gorilla.mux to correctly find the target router
	// as it uses the first matching and /api overrides the more accurate /api/worker
	if hasValue(enabledModules, "worker") {
		var workerRouter *mux.Router
		if err := container.Resolve(&workerRouter, di.Name("worker")); err != nil {
			return nil, err
		}

		mount(router, "/api/worker", workerRouter)
	}

	if hasValue(enabledModules, "api") {
		var apiRouter *mux.Router
		if err := container.Resolve(&apiRouter, di.Name("api")); err != nil {
			return nil, err
		}

		var authenticator Authenticator
		if err := container.Resolve(&authenticator); err != nil {
			return nil, err
		}

		apiRouter.Use(CreateAuthenticationMiddleware(authenticator))

		mount(router, "/api", apiRouter)
	}

	err := container.Invoke(enableReporters)
	if err != nil {
		return nil, err
	}

	// Resolve health checkers last, because all the services required by the application
	// must first be initialized and each of them can publish its own checkers
	var healthCheckers []*namedHealthChecker
	if container.Has(&healthCheckers) {
		if err := container.Resolve(&healthCheckers); err != nil {
			return nil, err
		}

		checkersOptions := make([]healthcheck.Option, len(healthCheckers))
		for i, checker := range healthCheckers {
			checkersOptions[i] = healthcheck.WithChecker(checker.Name, checker.Checker)
		}

		router.Handle("/healthcheck", healthcheck.Handler(checkersOptions...)).Methods("GET")
	}

	return router, nil
}

func newSkinsystemHandler(
	config *viper.Viper,
	emitter Emitter,
	skinsRepository SkinsRepository,
	capesRepository CapesRepository,
	mojangTexturesProvider MojangTexturesProvider,
	texturesSigner TexturesSigner,
) *mux.Router {
	config.SetDefault("textures.extra_param_name", "chrly")
	config.SetDefault("textures.extra_param_value", "how do you tame a horse in Minecraft?")

	return (&Skinsystem{
		Emitter:                 emitter,
		SkinsRepo:               skinsRepository,
		CapesRepo:               capesRepository,
		MojangTexturesProvider:  mojangTexturesProvider,
		TexturesSigner:          texturesSigner,
		TexturesExtraParamName:  config.GetString("textures.extra_param_name"),
		TexturesExtraParamValue: config.GetString("textures.extra_param_value"),
	}).Handler()
}

func newApiHandler(skinsRepository SkinsRepository) *mux.Router {
	return (&Api{
		SkinsRepo: skinsRepository,
	}).Handler()
}

func newUUIDsWorkerHandler(mojangUUIDsProvider *mojangtextures.BatchUuidsProvider) *mux.Router {
	return (&UUIDsWorker{
		MojangUuidsProvider: mojangUUIDsProvider,
	}).Handler()
}

func hasValue(slice []string, needle string) bool {
	for _, value := range slice {
		if value == needle {
			return true
		}
	}

	return false
}

func mount(router *mux.Router, path string, handler http.Handler) {
	router.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
}

type namedHealthChecker struct {
	Name    string
	Checker healthcheck.Checker
}
