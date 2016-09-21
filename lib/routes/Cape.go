package routes

import (
	"os"
	"io"
	"log"
	"strings"
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/tools"
	"elyby/minecraft-skinsystem/lib/services"
)

func Cape(response http.ResponseWriter, request *http.Request) {
	username := tools.ParseUsername(mux.Vars(request)["username"])
	log.Println("request cape for username " + username)
	file, err := os.Open(services.RootFolder + "/data/capes/" + strings.ToLower(username) + ".png")
	if (err != nil) {
		http.Redirect(response, request, "http://skins.minecraft.net/MinecraftCloaks/" + username + ".png", 301)
	}

	request.Header.Set("Content-Type", "image/png")
	io.Copy(response, file)
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
