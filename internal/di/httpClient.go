package di

import (
	"net/http"

	"github.com/defval/di"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var httpClientDiOptions = di.Options(
	di.Provide(newHttpClient),
)

func newHttpClient() *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
