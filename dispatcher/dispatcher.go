package dispatcher

import "github.com/asaskevich/EventBus"

type EventDispatcher interface {
	Subscribe(name string, fn interface{})
	Emit(name string, args ...interface{})
}

type LocalEventDispatcher struct {
	bus EventBus.Bus
}

func (d *LocalEventDispatcher) Subscribe(name string, fn interface{}) {
	_ = d.bus.Subscribe(name, fn)
}

func (d *LocalEventDispatcher) Emit(name string, args ...interface{}) {
	d.bus.Publish(name, args...)
}

func New() EventDispatcher {
	return &LocalEventDispatcher{
		bus: EventBus.New(),
	}
}
