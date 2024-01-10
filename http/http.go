package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/mono83/slf"
	"github.com/mono83/slf/wd"

	"github.com/elyby/chrly/dispatcher"
	v "github.com/elyby/chrly/version"
)

type Emitter interface {
	dispatcher.Emitter
}

func StartServer(server *http.Server, logger slf.Logger) {
	logger.Debug("Chrly :v (:c)", wd.StringParam("v", v.Version()), wd.StringParam("c", v.Commit()))

	done := make(chan bool, 1)
	go func() {
		logger.Info("Starting the server, HTTP on: :addr", wd.StringParam("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Emergency("Error in main(): :err", wd.ErrParam(err))
			close(done)
		}
	}()

	go func() {
		s := waitForExitSignal()
		logger.Info("Got signal: :signal, starting graceful shutdown", wd.StringParam("signal", s.String()))
		_ = server.Shutdown(context.Background())
		logger.Info("Graceful shutdown succeed, exiting", wd.StringParam("signal", s.String()))
		close(done)
	}()

	<-done
}

func waitForExitSignal() os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, os.Kill)

	return <-ch
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func CreateRequestEventsMiddleware(emitter Emitter, prefix string) mux.MiddlewareFunc {
	beforeTopic := strings.Join([]string{prefix, "before_request"}, ":")
	afterTopic := strings.Join([]string{prefix, "after_request"}, ":")

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			emitter.Emit(beforeTopic, req)

			lrw := &loggingResponseWriter{
				ResponseWriter: resp,
				statusCode:     http.StatusOK,
			}
			handler.ServeHTTP(lrw, req)

			emitter.Emit(afterTopic, req, lrw.statusCode)
		})
	}
}

type Authenticator interface {
	Authenticate(req *http.Request) error
}

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

func apiForbidden(resp http.ResponseWriter, reason string) {
	resp.WriteHeader(http.StatusForbidden)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(map[string]interface{}{
		"error": reason,
	})
	_, _ = resp.Write(result)
}

func apiNotFound(resp http.ResponseWriter, reason string) {
	resp.WriteHeader(http.StatusNotFound)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal([]interface{}{
		reason,
	})
	_, _ = resp.Write(result)
}
