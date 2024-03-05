package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mono83/slf"
	"github.com/mono83/slf/wd"

	"ely.by/chrly/internal/security"
)

func StartServer(ctx context.Context, server *http.Server, logger slf.Logger) {
	srvErr := make(chan error, 1)
	go func() {
		logger.Info("Starting the server, HTTP on: :addr", wd.StringParam("addr", server.Addr))
		srvErr <- server.ListenAndServe()
		close(srvErr)
	}()

	select {
	case err := <-srvErr:
		logger.Emergency("Error in main(): :err", wd.ErrParam(err))
	case <-ctx.Done():
		logger.Info("Got stop signal, starting graceful shutdown: :ctx")

		stopCtx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelFunc()

		_ = server.Shutdown(stopCtx)

		logger.Info("Graceful shutdown succeed, exiting")
	}
}

type Authenticator interface {
	Authenticate(req *http.Request, scope security.Scope) error
}

func NewAuthenticationMiddleware(authenticator Authenticator, scope security.Scope) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			err := authenticator.Authenticate(req, scope)
			if err != nil {
				apiForbidden(resp, err.Error())
				return
			}

			handler.ServeHTTP(resp, req)
		})
	}
}

func NewConditionalMiddleware(cond func(req *http.Request) bool, m mux.MiddlewareFunc) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if cond(req) {
				handler = m.Middleware(handler)
			}

			handler.ServeHTTP(resp, req)
		})
	}
}

func NotFoundHandler(response http.ResponseWriter, _ *http.Request) {
	data, _ := json.Marshal(map[string]string{
		"status":  "404",
		"message": "Not Found",
	})

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	_, _ = response.Write(data)
}

func apiBadRequest(resp http.ResponseWriter, errorsPerField map[string][]string) {
	resp.WriteHeader(http.StatusBadRequest)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(map[string]any{
		"errors": errorsPerField,
	})
	_, _ = resp.Write(result)
}

var internalServerError = []byte("Internal server error")

func apiServerError(resp http.ResponseWriter, err error) {
	resp.WriteHeader(http.StatusInternalServerError)
	resp.Header().Set("Content-Type", "text/plain")
	_, _ = resp.Write(internalServerError)
}

func apiForbidden(resp http.ResponseWriter, reason string) {
	resp.WriteHeader(http.StatusForbidden)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(map[string]any{
		"error": reason,
	})
	_, _ = resp.Write(result)
}
