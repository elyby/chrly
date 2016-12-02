package routes

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/tools"
	"elyby/minecraft-skinsystem/lib/data"
	"elyby/minecraft-skinsystem/lib/services"
)

func Cape(response http.ResponseWriter, request *http.Request) {
	if (mux.Vars(request)["converted"] == "") {
		services.Logger.IncCounter("capes.request", 1)
	}

	username := tools.ParseUsername(mux.Vars(request)["username"])
	rec, err := data.FindCapeByUsername(username)
	if (err != nil) {
		http.Redirect(response, request, "http://skins.minecraft.net/MinecraftCloaks/" + username + ".png", 301)
	}

	request.Header.Set("Content-Type", "image/png")
	io.Copy(response, rec.File)
}

func CapeGET(w http.ResponseWriter, r *http.Request) {
	services.Logger.IncCounter("capes.get-request", 1)
	username := r.URL.Query().Get("name")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(r)["username"] = username
	mux.Vars(r)["converted"] = "1"
	Cape(w, r)
}
