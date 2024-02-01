package eventsubscribers

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/etherlabsio/healthcheck/v2"
)

type Pingable interface {
	Ping() error
}

func DatabaseChecker(connection Pingable) healthcheck.CheckerFunc {
	return func(ctx context.Context) error {
		done := make(chan error)
		go func() {
			done <- connection.Ping()
		}()

		select {
		case <-ctx.Done():
			return errors.New("check timeout")
		case err := <-done:
			return err
		}
	}
}

type expiringErrHolder struct {
	D   time.Duration
	err error
	l   sync.Mutex
	t   *time.Timer
}

func (h *expiringErrHolder) Get() error {
	h.l.Lock()
	defer h.l.Unlock()

	return h.err
}

func (h *expiringErrHolder) Set(err error) {
	h.l.Lock()
	defer h.l.Unlock()
	if h.t != nil {
		h.t.Stop()
		h.t = nil
	}

	h.err = err
	if err != nil {
		h.t = time.AfterFunc(h.D, func() {
			h.Set(nil)
		})
	}
}
