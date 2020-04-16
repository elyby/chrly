package di

import (
	"github.com/goava/di"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/http"
)

var handlers = di.Options(
	di.Provide(newSkinsystemHandler, di.WithName("skinsystem")),
)

func newSkinsystemHandler(
	config *viper.Viper,
	emitter http.Emitter,
	skinsRepository http.SkinsRepository,
	capesRepository http.CapesRepository,
	mojangTexturesProvider http.MojangTexturesProvider,
) *mux.Router {
	handlerFactory := &http.Skinsystem{
		Emitter:                 emitter,
		SkinsRepo:               skinsRepository,
		CapesRepo:               capesRepository,
		MojangTexturesProvider:  mojangTexturesProvider,
		TexturesExtraParamName:  config.GetString("textures.extra_param_name"),
		TexturesExtraParamValue: config.GetString("textures.extra_param_value"),
	}

	return handlerFactory.CreateHandler()
}

// TODO: pin implementation to make it non-configurable
func newUUIDsWorkerHandler(mojangUUIDsProvider http.MojangUuidsProvider) *mux.Router {
	handlerFactory := &http.UUIDsWorker{
		UUIDsProvider: mojangUUIDsProvider,
	}

	return handlerFactory.CreateHandler()
}
