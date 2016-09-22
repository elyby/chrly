package routes

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/tools"
	"elyby/minecraft-skinsystem/lib/data"
)

func Cape(response http.ResponseWriter, request *http.Request) {
	username := tools.ParseUsername(mux.Vars(request)["username"])
	log.Println("request cape for username " + username)
	rec, err := data.FindCapeByUsername(username)
	if (err != nil) {
		http.Redirect(response, request, "http://skins.minecraft.net/MinecraftCloaks/" + username + ".png", 301)
	}

	request.Header.Set("Content-Type", "image/png")
	io.Copy(response, rec.File)
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
