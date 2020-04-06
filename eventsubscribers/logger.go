package eventsubscribers

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"

	"github.com/mono83/slf"
	"github.com/mono83/slf/wd"

	"github.com/elyby/chrly/api/mojang"
)

type Logger struct {
	slf.Logger
}

func (l *Logger) ConfigureWithDispatcher(d Subscriber) {
	d.Subscribe("skinsystem:after_request", l.handleAfterSkinsystemRequest)

	d.Subscribe("mojang_textures:usernames:after_call", l.createMojangTexturesErrorHandler("usernames"))
	d.Subscribe("mojang_textures:textures:after_call", l.createMojangTexturesErrorHandler("textures"))
}

func (l *Logger) handleAfterSkinsystemRequest(req *http.Request, statusCode int) {
	path := req.URL.Path
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}

	l.Info(
		":ip - - \":method :path\" :statusCode - \":userAgent\" \":forwardedIp\"",
		wd.StringParam("ip", trimPort(req.RemoteAddr)),
		wd.StringParam("method", req.Method),
		wd.StringParam("path", path),
		wd.IntParam("statusCode", statusCode),
		wd.StringParam("userAgent", req.UserAgent()),
		wd.StringParam("forwardedIp", req.Header.Get("X-Forwarded-For")),
	)
}

func (l *Logger) createMojangTexturesErrorHandler(provider string) func(identity string, result interface{}, err error) {
	providerParam := wd.NameParam(provider)
	return func(identity string, result interface{}, err error) {
		if err == nil {
			return
		}

		errParam := wd.ErrParam(err)

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

func trimPort(ip string) string {
	// Don't care about possible -1 result because RemoteAddr will always contain ip and port
	cutTo := strings.LastIndexByte(ip, ':')

	return ip[0:cutTo]
}
