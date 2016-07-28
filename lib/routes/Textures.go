package routes

import (
	"log"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/data"
	"elyby/minecraft-skinsystem/lib/tools"
)

func Textures(w http.ResponseWriter, r *http.Request) {
	username := tools.ParseUsername(mux.Vars(r)["username"])
	log.Println("request textures for username " + username)

	rec, err := data.FindRecord(username)
	if (err != nil || rec.SkinId == 0) {
		rec.Url = "http://skins.minecraft.net/MinecraftSkins/" + username + ".png"
		rec.Hash = string(tools.BuildNonElyTexturesHash(username))
	}

	textures := data.TexturesResponse{
		Skin: &data.Skin{
			Url: tools.BuildElyUrl(rec.Url),
			Hash: rec.Hash,
		},
	}

	if (rec.IsSlim) {
		textures.Skin.Metadata = &data.SkinMetadata{
			Model: "slim",
		}
	}

	response,_ := json.Marshal(textures)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
