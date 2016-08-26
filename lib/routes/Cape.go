package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/tools"
)

func Cape(w http.ResponseWriter, r *http.Request) {
	username := tools.ParseUsername(mux.Vars(r)["username"])
	log.Println("request cape for username " + username)
	http.Redirect(w, r, "http://skins.minecraft.net/MinecraftCloaks/" + username + ".png", 301)
}

func CapeGET(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("name")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(r)["username"] = username
	Cape(w, r)
}
