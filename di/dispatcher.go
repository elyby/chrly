package di

import (
	"github.com/goava/di"
	"github.com/mono83/slf"

	d "github.com/elyby/chrly/dispatcher"
	"github.com/elyby/chrly/eventsubscribers"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/mojangtextures"
)

var dispatcher = di.Options(
	di.Provide(newDispatcher,
		di.As(new(http.Emitter)),
		di.As(new(mojangtextures.Emitter)),
		di.As(new(eventsubscribers.Subscriber)),
	),
	di.Invoke(enableEventsHandlers),
)

func newDispatcher() d.EventDispatcher {
	return d.New()
}

func enableEventsHandlers(
	dispatcher d.EventDispatcher,
	logger slf.Logger,
	statsReporter slf.StatsReporter,
) {
	// TODO: use idea from https://github.com/goava/di/issues/10#issuecomment-615869852
	(&eventsubscribers.Logger{Logger: logger}).ConfigureWithDispatcher(dispatcher)
	(&eventsubscribers.StatsReporter{StatsReporter: statsReporter}).ConfigureWithDispatcher(dispatcher)
}
