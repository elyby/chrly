package http

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Emitter interface {
	Emit(name string, args ...interface{})
}

func Serve(address string, handler http.Handler) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	server := &http.Server{
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 16,
		Handler:        handler,
	}

	return server.Serve(listener)
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

func NotFound(response http.ResponseWriter, _ *http.Request) {
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

func apiServerError(resp http.ResponseWriter) {
	resp.WriteHeader(http.StatusInternalServerError)
}
