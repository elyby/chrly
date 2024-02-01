package di

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/defval/di"
	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"

	. "ely.by/chrly/internal/http"
	"ely.by/chrly/internal/security"
)

var server = di.Options(
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

type serverParams struct {
	di.Inject

	Config  *viper.Viper  `di:""`
	Handler http.Handler  `di:""`
	Sentry  *raven.Client `di:"" optional:"true"`
}

func newServer(params serverParams) *http.Server {
	params.Config.SetDefault("server.host", "")
	params.Config.SetDefault("server.port", 80)

	var handler http.Handler
	if params.Sentry != nil {
		// raven.Recoverer uses DefaultClient and nothing can be done about it
		// To avoid code duplication, if the Sentry service is successfully initiated,
		// it will also replace DefaultClient, so raven.Recoverer will work with the instance
		// created in the application constructor
		handler = raven.Recoverer(params.Handler)
	} else {
		// Raven's Recoverer is prints the stacktrace and sets the corresponding status itself.
		// But there is no magic and if you don't define a panic handler, Mux will just reset the connection
		handler = http.HandlerFunc(func(request http.ResponseWriter, response *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					debug.PrintStack() // TODO: colorize output
					request.WriteHeader(http.StatusInternalServerError)
				}
			}()

			params.Handler.ServeHTTP(request, response)
		})
	}

	address := fmt.Sprintf("%s:%d", params.Config.GetString("server.host"), params.Config.GetInt("server.port"))
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
