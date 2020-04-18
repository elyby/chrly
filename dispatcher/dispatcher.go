package dispatcher

import "github.com/asaskevich/EventBus"

// TODO: split on 2 interfaces and use them across the application
type EventDispatcher interface {
	Subscribe(topic string, fn interface{})
	Emit(topic string, args ...interface{})
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

func New() EventDispatcher {
	return &localEventDispatcher{
		bus: EventBus.New(),
	}
}
