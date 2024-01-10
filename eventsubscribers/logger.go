package eventsubscribers

import (
	"net/http"
	"strings"

	"github.com/mono83/slf"
	"github.com/mono83/slf/wd"
)

type Logger struct {
	slf.Logger
}

func (l *Logger) ConfigureWithDispatcher(d Subscriber) {
	d.Subscribe("skinsystem:after_request", l.handleAfterSkinsystemRequest)
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

func trimPort(ip string) string {
	// Don't care about possible -1 result because RemoteAddr will always contain ip and port
	cutTo := strings.LastIndexByte(ip, ':')

	return ip[0:cutTo]
}
