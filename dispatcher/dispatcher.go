package dispatcher

import "github.com/asaskevich/EventBus"

type Subscriber interface {
	Subscribe(topic string, fn interface{})
}

type Emitter interface {
	Emit(topic string, args ...interface{})
}

type Dispatcher interface {
	Subscriber
	Emitter
}

type localEventDispatcher struct {
	bus EventBus.Bus
}

func (d *localEventDispatcher) Subscribe(topic string, fn interface{}) {
	_ = d.bus.Subscribe(topic, fn)
}

func (d *localEventDispatcher) Emit(topic string, args ...interface{}) {
	d.bus.Publish(topic, args...)
}

func New() Dispatcher {
	return &localEventDispatcher{
		bus: EventBus.New(),
	}
}
