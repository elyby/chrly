package http

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func (cfg *Config) Cape(response http.ResponseWriter, request *http.Request) {
	if mux.Vars(request)["converted"] == "" {
		cfg.Logger.IncCounter("capes.request", 1)
	}

	username := parseUsername(mux.Vars(request)["username"])
	rec, err := cfg.CapesRepo.FindByUsername(username)
	if err == nil {
		request.Header.Set("Content-Type", "image/png")
		io.Copy(response, rec.File)
		return
	}

	mojangTextures := <-cfg.MojangTexturesQueue.GetTexturesForUsername(username)
	if mojangTextures == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	texturesProp := mojangTextures.DecodeTextures()
	cape := texturesProp.Textures.Cape
	if cape == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(response, request, cape.Url, 301)
}

func (cfg *Config) CapeGET(response http.ResponseWriter, request *http.Request) {
	cfg.Logger.IncCounter("capes.get_request", 1)
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username
	mux.Vars(request)["converted"] = "1"

	cfg.Cape(response, request)
}
