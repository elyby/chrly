package ui

import (
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/utils"
)

func (s *uiService) Skin(response http.ResponseWriter, request *http.Request) {
	if mux.Vars(request)["converted"] == "" {
		s.logger.IncCounter("skins.request", 1)
	}

	username := utils.ParseUsername(mux.Vars(request)["username"])
	rec, err := s.skinsRepo.FindByUsername(username)
	if err != nil {
		http.Redirect(response, request, "http://skins.minecraft.net/MinecraftSkins/" + username + ".png", 301)
		return
	}

	http.Redirect(response, request, utils.BuildElyUrl(rec.Url), 301)
}

func (s *uiService) SkinGET(response http.ResponseWriter, request *http.Request) {
	s.logger.IncCounter("skins.get_request", 1)
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username
	mux.Vars(request)["converted"] = "1"

	s.Skin(response, request)
}
