package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/tools"
	"elyby/minecraft-skinsystem/lib/data"
	"elyby/minecraft-skinsystem/lib/services"
)

func Skin(w http.ResponseWriter, r *http.Request) {
	if (mux.Vars(r)["converted"] == "") {
		services.Logger.IncCounter("skins.request", 1)
	}

	username := tools.ParseUsername(mux.Vars(r)["username"])
	rec, err := data.FindSkinByUsername(username)
	if (err != nil) {
		http.Redirect(w, r, "http://skins.minecraft.net/MinecraftSkins/" + username + ".png", 301)
		return
	}

	http.Redirect(w, r, tools.BuildElyUrl(rec.Url), 301);
}

func SkinGET(w http.ResponseWriter, r *http.Request) {
	services.Logger.IncCounter("skins.get-request", 1)
	username := r.URL.Query().Get("name")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(r)["username"] = username
	mux.Vars(r)["converted"] = "1"
	Skin(w, r)
}
