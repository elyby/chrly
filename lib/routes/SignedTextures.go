package routes

import (
	"strings"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/data"
	"elyby/minecraft-skinsystem/lib/tools"
	"elyby/minecraft-skinsystem/lib/services"
)

func SignedTextures(w http.ResponseWriter, r *http.Request) {
	services.Logger.IncCounter("signed_textures.request", 1)
	username := tools.ParseUsername(mux.Vars(r)["username"])

	rec, err := data.FindSkinByUsername(username)
	if (err != nil || rec.SkinId == 0 || rec.MojangTextures == "") {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	responseData:= data.SignedTexturesResponse{
		Id: strings.Replace(rec.Uuid, "-", "", -1),
		Name: rec.Username,
		Props: []data.Property{
			{
				Name: "textures",
				Signature: rec.MojangSignature,
				Value: rec.MojangTextures,
			},
			{
				Name: "ely",
				Value: "but why are you asking?",
			},
		},
	}

	response,_ := json.Marshal(responseData)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
