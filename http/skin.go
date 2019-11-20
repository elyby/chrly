package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (cfg *Config) Skin(response http.ResponseWriter, request *http.Request) {
	if mux.Vars(request)["converted"] == "" {
		cfg.Logger.IncCounter("skins.request", 1)
	}

	username := parseUsername(mux.Vars(request)["username"])
	rec, err := cfg.SkinsRepo.FindByUsername(username)
	if err == nil && rec.SkinId != 0 {
		http.Redirect(response, request, rec.Url, 301)
		return
	}

	mojangTextures, err := cfg.MojangTexturesProvider.GetForUsername(username)
	if err != nil || mojangTextures == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	texturesProp := mojangTextures.DecodeTextures()
	skin := texturesProp.Textures.Skin
	if skin == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(response, request, skin.Url, 301)
}

func (cfg *Config) SkinGET(response http.ResponseWriter, request *http.Request) {
	cfg.Logger.IncCounter("skins.get_request", 1)
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username
	mux.Vars(request)["converted"] = "1"

	cfg.Skin(response, request)
}
