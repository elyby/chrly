package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/tools"
	"elyby/minecraft-skinsystem/lib/data"
)

const defaultHash = "default"

func Face(w http.ResponseWriter, r *http.Request) {
	username := tools.ParseUsername(mux.Vars(r)["username"])
	log.Println("request skin for username " + username);
	rec, err := data.FindSkinByUsername(username)
	var hash string
	if (err != nil || rec.SkinId == 0) {
		hash = defaultHash;
	} else {
		hash = rec.Hash
	}

	http.Redirect(w, r, tools.BuildElyUrl(buildFaceUrl(hash)), 301);
}

func buildFaceUrl(hash string) string {
	return "/minecraft/skin_buffer/faces/" + hash + ".png"
}
