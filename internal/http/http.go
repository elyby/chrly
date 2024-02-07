package http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/mono83/slf"
	"github.com/mono83/slf/wd"

	"ely.by/chrly/internal/version"
)

func StartServer(server *http.Server, logger slf.Logger) {
	logger.Debug("Chrly :v (:c)", wd.StringParam("v", version.Version()), wd.StringParam("c", version.Commit()))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, os.Kill)
	defer cancel()

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
	Authenticate(req *http.Request) error
}

// The current middleware implementation doesn't check the scope assigned to the token.
// For now there is only one scope and at this moment I don't want to spend time on it
func CreateAuthenticationMiddleware(checker Authenticator) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			err := checker.Authenticate(req)
			if err != nil {
				apiForbidden(resp, err.Error())
				return
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
	result, _ := json.Marshal(map[string]interface{}{
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
	result, _ := json.Marshal(map[string]interface{}{
		"error": reason,
	})
	_, _ = resp.Write(result)
}
