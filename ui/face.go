package ui

import (
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/utils"
)

const defaultHash = "default"

func (s *uiService) Face(response http.ResponseWriter, request *http.Request) {
	username := utils.ParseUsername(mux.Vars(request)["username"])
	rec, err := s.skinsRepo.FindByUsername(username)
	var hash string
	if err != nil || rec.SkinId == 0 {
		hash = defaultHash
	} else {
		hash = rec.Hash
	}

	http.Redirect(response, request, utils.BuildElyUrl(buildFaceUrl(hash)), 301)
}

func buildFaceUrl(hash string) string {
	return "/minecraft/skin_buffer/faces/" + hash + ".png"
}
