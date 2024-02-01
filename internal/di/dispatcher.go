package di

import (
	"github.com/defval/di"
	"github.com/mono83/slf"

	d "github.com/elyby/chrly/internal/dispatcher"
	"github.com/elyby/chrly/internal/eventsubscribers"
	"github.com/elyby/chrly/internal/http"
)

var dispatcher = di.Options(
	di.Provide(newDispatcher,
		di.As(new(d.Emitter)),
		di.As(new(d.Subscriber)),
		di.As(new(http.Emitter)),
		di.As(new(eventsubscribers.Subscriber)),
	),
	di.Invoke(enableEventsHandlers),
)

func newDispatcher() d.Dispatcher {
	return d.New()
}

func enableEventsHandlers(
	dispatcher d.Subscriber,
	logger slf.Logger,
	statsReporter slf.StatsReporter,
) {
	// TODO: use idea from https://github.com/defval/di/issues/10#issuecomment-615869852
	(&eventsubscribers.Logger{Logger: logger}).ConfigureWithDispatcher(dispatcher)
	(&eventsubscribers.StatsReporter{StatsReporter: statsReporter}).ConfigureWithDispatcher(dispatcher)
}
