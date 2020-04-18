package di

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/goava/di"
	"github.com/spf13/viper"

	. "github.com/elyby/chrly/http"
)

var server = di.Options(
	di.Provide(newAuthenticator, di.As(new(Authenticator))),
	di.Provide(newServer),
)

func newAuthenticator(config *viper.Viper, emitter Emitter) (*JwtAuth, error) {
	key := config.GetString("chrly.secret")
	if key == "" {
		return nil, errors.New("chrly.secret must be set in order to use authenticator")
	}

	return &JwtAuth{
		Key:     []byte(key),
		Emitter: emitter,
	}, nil
}

func newServer(config *viper.Viper, handler http.Handler) *http.Server {
	config.SetDefault("server.host", "")
	config.SetDefault("server.port", 80)

	address := fmt.Sprintf("%s:%d", config.GetString("server.host"), config.GetInt("server.port"))
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
