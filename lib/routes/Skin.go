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

	http.Redirect(w, r, rec.Url, 301);
}
