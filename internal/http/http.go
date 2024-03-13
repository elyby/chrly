package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"ely.by/chrly/internal/security"
)

func StartServer(ctx context.Context, server *http.Server) {
	srvErr := make(chan error, 1)
	go func() {
		slog.Info("Starting the server", slog.String("addr", server.Addr))
		srvErr <- server.ListenAndServe()
		close(srvErr)
	}()

	select {
	case err := <-srvErr:
		slog.Error("Error in the server", slog.Any("error", err))
	case <-ctx.Done():
		slog.Info("Got stop signal, starting graceful shutdown")

		stopCtx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelFunc()

		_ = server.Shutdown(stopCtx)

		slog.Info("Graceful shutdown succeed, exiting")
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

func apiServerError(resp http.ResponseWriter, req *http.Request, err error) {
	span := trace.SpanFromContext(req.Context())
	span.SetStatus(codes.Error, "")
	span.RecordError(err)

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
