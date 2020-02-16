package dispatcher

import "github.com/asaskevich/EventBus"

type EventDispatcher interface {
	Subscribe(topic string, fn interface{})
	Emit(topic string, args ...interface{})
}

type LocalEventDispatcher struct {
	bus EventBus.Bus
}

func (d *LocalEventDispatcher) Subscribe(topic string, fn interface{}) {
	_ = d.bus.Subscribe(topic, fn)
}

func (d *LocalEventDispatcher) Emit(topic string, args ...interface{}) {
	d.bus.Publish(topic, args...)
}

func New() EventDispatcher {
	return &LocalEventDispatcher{
		bus: EventBus.New(),
	}
}
