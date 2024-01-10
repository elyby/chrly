package eventsubscribers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type pingableMock struct {
	mock.Mock
}

func (p *pingableMock) Ping() error {
	args := p.Called()
	return args.Error(0)
}

func TestDatabaseChecker(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		p := &pingableMock{}
		p.On("Ping").Return(nil)
		checker := DatabaseChecker(p)
		assert.Nil(t, checker(context.Background()))
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("mock error")
		p := &pingableMock{}
		p.On("Ping").Return(err)
		checker := DatabaseChecker(p)
		assert.Equal(t, err, checker(context.Background()))
	})

	t.Run("context timeout", func(t *testing.T) {
		p := &pingableMock{}
		waitChan := make(chan time.Time, 1)
		p.On("Ping").WaitUntil(waitChan).Return(nil)

		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()

		checker := DatabaseChecker(p)
		assert.Errorf(t, checker(ctx), "check timeout")
		close(waitChan)
	})
}
