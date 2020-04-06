package eventsubscribers

type Subscriber interface {
	Subscribe(topic string, fn interface{})
}
