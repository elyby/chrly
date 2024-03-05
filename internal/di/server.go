package di

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/defval/di"
	"github.com/spf13/viper"

	. "ely.by/chrly/internal/http"
	"ely.by/chrly/internal/security"
)

var serverDiOptions = di.Options(
	di.Provide(newAuthenticator, di.As(new(Authenticator))),
	di.Provide(newServer),
)

func newAuthenticator(config *viper.Viper) (*security.Jwt, error) {
	key := config.GetString("chrly.secret")
	if key == "" {
		return nil, errors.New("chrly.secret must be set in order to use authenticator")
	}

	return security.NewJwt([]byte(key)), nil
}

func newServer(Config *viper.Viper, Handler http.Handler) *http.Server {
	Config.SetDefault("server.host", "")
	Config.SetDefault("server.port", 80)

	var handler http.Handler = http.HandlerFunc(func(request http.ResponseWriter, response *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				debug.PrintStack()
				request.WriteHeader(http.StatusInternalServerError)
			}
		}()

		Handler.ServeHTTP(request, response)
	})

	address := fmt.Sprintf("%s:%d", Config.GetString("server.host"), Config.GetInt("server.port"))
	server := &http.Server{
		Addr:           address,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 16,
		Handler:        handler,
	}

	return server
}
