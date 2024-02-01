package eventsubscribers

import "github.com/elyby/chrly/internal/dispatcher"

type Subscriber interface {
	dispatcher.Subscriber
}
