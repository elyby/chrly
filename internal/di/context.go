package di

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/defval/di"
)

var contextDiOptions = di.Options(
	di.Provide(newBaseContext),
)

func newBaseContext() context.Context {
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, os.Kill)

	return ctx
}
