package eventsubscribers

import (
	"net"
	"net/url"
	"syscall"

	"github.com/mono83/slf"
	"github.com/mono83/slf/wd"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/dispatcher"
)

type Logger struct {
	slf.Logger
}

func (l *Logger) ConfigureWithDispatcher(d dispatcher.EventDispatcher) {
	d.Subscribe("mojang_textures:usernames:after_call", l.createMojangTexturesErrorHandler("usernames"))
	d.Subscribe("mojang_textures:textures:after_call", l.createMojangTexturesErrorHandler("textures"))
}

func (l *Logger) createMojangTexturesErrorHandler(provider string) func(identity string, result interface{}, err error) {
	providerParam := wd.NameParam(provider)
	return func(identity string, result interface{}, err error) {
		if err == nil {
			return
		}

		errParam := wd.ErrParam(err)

		l.Debug("Mojang "+provider+" provider resulted an error :err", errParam)

		switch err.(type) {
		case *mojang.BadRequestError:
			l.logMojangTexturesWarning(providerParam, errParam)
			return
		case *mojang.ForbiddenError:
			l.logMojangTexturesWarning(providerParam, errParam)
			return
		case *mojang.TooManyRequestsError:
			l.logMojangTexturesWarning(providerParam, errParam)
			return
		case net.Error:
			if err.(net.Error).Timeout() {
				return
			}

			if _, ok := err.(*url.Error); ok {
				return
			}

			if opErr, ok := err.(*net.OpError); ok && (opErr.Op == "dial" || opErr.Op == "read") {
				return
			}

			if err == syscall.ECONNREFUSED {
				return
			}
		}

		l.Error(":name: Unexpected Mojang response error: :err", providerParam, errParam)
	}
}

func (l *Logger) logMojangTexturesWarning(providerParam slf.Param, errParam slf.Param) {
	l.Warning(":name: :err", providerParam, errParam)
}
