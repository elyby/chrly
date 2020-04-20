package eventsubscribers

import "github.com/elyby/chrly/dispatcher"

type Subscriber interface {
	dispatcher.Subscriber
}
