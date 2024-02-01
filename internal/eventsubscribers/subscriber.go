package eventsubscribers

import "ely.by/chrly/internal/dispatcher"

type Subscriber interface {
	dispatcher.Subscriber
}
