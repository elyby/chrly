package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

const defaultHash = "default"

func (cfg *Config) Face(response http.ResponseWriter, request *http.Request) {
	cfg.Logger.IncCounter("faces.request", 1)
	username := parseUsername(mux.Vars(request)["username"])
	rec, err := cfg.SkinsRepo.FindByUsername(username)
	var hash string
	if err != nil || rec.SkinId == 0 {
		hash = defaultHash
	} else {
		hash = rec.Hash
	}

	http.Redirect(response, request, buildFaceUrl(hash), 301)
}

func buildFaceUrl(hash string) string {
	return "http://ely.by/minecraft/skin_buffer/faces/" + hash + ".png"
}
