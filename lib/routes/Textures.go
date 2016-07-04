package routes

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"elyby/minecraft-skinsystem/lib/structures"
	"encoding/json"
	"elyby/minecraft-skinsystem/lib/tools"
)

func GetTextures(w http.ResponseWriter, r *http.Request) {
	username := tools.ParseUsername(mux.Vars(r)["username"])
	log.Println("request textures for username " + username)

	rec, err := tools.FindRecord(username)
	if (err != nil || rec.SkinId == 0) {
		rec.Url = "http://skins.minecraft.net/MinecraftSkins/" + username + ".png"
		rec.Hash = string(tools.BuildNonElyTexturesHash(username))
	}

	textures := structures.TexturesResponse{
		Skin: &structures.Skin{
			Url: rec.Url,
			Hash: rec.Hash,
		},
	}

	if (rec.IsSlim) {
		textures.Skin.Metadata = &structures.SkinMetadata{
			Model: "slim",
		}
	}

	response,_ := json.Marshal(textures)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
