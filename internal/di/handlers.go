package di

import (
	"net/http"
	"slices"
	"strings"

	"github.com/defval/di"
	"github.com/etherlabsio/healthcheck/v2"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	. "ely.by/chrly/internal/http"
	"ely.by/chrly/internal/security"
)

const ModuleSkinsystem = "skinsystem"
const ModuleProfiles = "profiles"
const ModuleSigner = "signer"

var handlersDiOptions = di.Options(
	di.Provide(newHandlerFactory, di.As(new(http.Handler))),
	di.Provide(newSkinsystemHandler, di.WithName(ModuleSkinsystem)),
	di.Provide(newProfilesApiHandler, di.WithName(ModuleProfiles)),
	di.Provide(newSignerApiHandler, di.WithName(ModuleSigner)),
)

func newHandlerFactory(
	container *di.Container,
	config *viper.Viper,
) (*mux.Router, error) {
	enabledModules := config.GetStringSlice("modules")

	// gorilla.mux has no native way to combine multiple routers.
	// The hack used later in the code works for prefixes in addresses, but leads to misbehavior
	// if you set an empty prefix. Since the main application should be mounted at the root prefix,
	// we use it as the base router
	var router *mux.Router
	if slices.Contains(enabledModules, ModuleSkinsystem) {
		if err := container.Resolve(&router, di.Name(ModuleSkinsystem)); err != nil {
			return nil, err
		}
	} else {
		router = mux.NewRouter()
	}

	router.StrictSlash(true)
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	if slices.Contains(enabledModules, ModuleProfiles) {
		var profilesApiRouter *mux.Router
		if err := container.Resolve(&profilesApiRouter, di.Name(ModuleProfiles)); err != nil {
			return nil, err
		}

		var authenticator Authenticator
		if err := container.Resolve(&authenticator); err != nil {
			return nil, err
		}

		profilesApiRouter.Use(NewAuthenticationMiddleware(authenticator, security.ProfilesScope))

		mount(router, "/api/profiles", profilesApiRouter)
	}

	if slices.Contains(enabledModules, ModuleSigner) {
		var signerApiRouter *mux.Router
		if err := container.Resolve(&signerApiRouter, di.Name(ModuleSigner)); err != nil {
			return nil, err
		}

		var authenticator Authenticator
		if err := container.Resolve(&authenticator); err != nil {
			return nil, err
		}

		authMiddleware := NewAuthenticationMiddleware(authenticator, security.SignScope)
		conditionalAuth := NewConditionalMiddleware(func(req *http.Request) bool {
			return req.Method != "GET"
		}, authMiddleware)
		signerApiRouter.Use(conditionalAuth)

		mount(router, "/api/signer", signerApiRouter)
	}

	// Resolve health checkers last, because all the services required by the application
	// must first be initialized and each of them can publish its own checkers
	var healthCheckers []*namedHealthChecker
	if has, _ := container.Has(&healthCheckers); has {
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
	profilesProvider ProfilesProvider,
	texturesSigner SignerService,
) *mux.Router {
	config.SetDefault("textures.extra_param_name", "chrly")
	config.SetDefault("textures.extra_param_value", "how do you tame a horse in Minecraft?")

	return (&Skinsystem{
		ProfilesProvider:        profilesProvider,
		SignerService:           texturesSigner,
		TexturesExtraParamName:  config.GetString("textures.extra_param_name"),
		TexturesExtraParamValue: config.GetString("textures.extra_param_value"),
	}).Handler()
}

func newProfilesApiHandler(profilesManager ProfilesManager) *mux.Router {
	return (&ProfilesApi{
		ProfilesManager: profilesManager,
	}).Handler()
}

func newSignerApiHandler(signer Signer) *mux.Router {
	return (&SignerApi{
		Signer: signer,
	}).Handler()
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