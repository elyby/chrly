package di

import (
	"github.com/goava/di"

	dispatcherModule "github.com/elyby/chrly/dispatcher"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

var dispatcher = di.Options(
	di.Provide(newDispatcher,
		di.As(new(http.Emitter)),
		di.As(new(mojangtextures.Emitter)),
	),
)

func newDispatcher() dispatcherModule.EventDispatcher {
	return dispatcherModule.New()
}
