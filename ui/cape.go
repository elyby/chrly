package ui

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/utils"
)

func (s *uiService) Cape(response http.ResponseWriter, request *http.Request) {
	if mux.Vars(request)["converted"] == "" {
		s.logger.IncCounter("capes.request", 1)
	}

	username := utils.ParseUsername(mux.Vars(request)["username"])
	rec, err := s.capesRepo.FindByUsername(username)
	if err != nil {
		http.Redirect(response, request, "http://skins.minecraft.net/MinecraftCloaks/" + username + ".png", 301)
	}

	request.Header.Set("Content-Type", "image/png")
	io.Copy(response, rec.File)
}

func (s *uiService) CapeGET(response http.ResponseWriter, request *http.Request) {
	s.logger.IncCounter("capes.get_request", 1)
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username
	mux.Vars(request)["converted"] = "1"

	s.Cape(response, request)
}
