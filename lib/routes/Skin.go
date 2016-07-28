package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/tools"
	"elyby/minecraft-skinsystem/lib/data"
)

func Skin(w http.ResponseWriter, r *http.Request) {
	username := tools.ParseUsername(mux.Vars(r)["username"])
	log.Println("request skin for username " + username);
	rec, err := data.FindRecord(username)
	if (err != nil) {
		http.Redirect(w, r, "http://skins.minecraft.net/MinecraftSkins/" + username + ".png", 301)
		log.Println("Cannot get skin for username " + username)
		return
	}

	http.Redirect(w, r, tools.BuildElyUrl(rec.Url), 301);
}

func SkinGET(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("name")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(r)["username"] = username
	Skin(w, r)
}
