package di

import (
	"net/http"

	"github.com/defval/di"
)

var httpClientDiOptions = di.Options(
	di.Provide(newHttpClient),
)

func newHttpClient() *http.Client {
	return &http.Client{}
}
