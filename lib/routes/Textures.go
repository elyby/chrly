package routes

import (
	"log"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/data"
	"elyby/minecraft-skinsystem/lib/tools"
	"elyby/minecraft-skinsystem/lib/services"
)

func Textures(w http.ResponseWriter, r *http.Request) {
	services.Logger.IncCounter("textures.request", 1)
	username := tools.ParseUsername(mux.Vars(r)["username"])

	rec, err := data.FindSkinByUsername(username)
	if (err != nil || rec.SkinId == 0) {
		rec.Url = "http://skins.minecraft.net/MinecraftSkins/" + username + ".png"
		rec.Hash = string(tools.BuildNonElyTexturesHash(username))
	} else {
		rec.Url = tools.BuildElyUrl(rec.Url)
	}

	textures := data.TexturesResponse{
		Skin: &data.Skin{
			Url: rec.Url,
			Hash: rec.Hash,
		},
	}

	if (rec.IsSlim) {
		textures.Skin.Metadata = &data.SkinMetadata{
			Model: "slim",
		}
	}

	capeRec, err := data.FindCapeByUsername(username)
	if (err == nil) {
		capeUrl, err := services.Router.Get("cloaks").URL("username", username)
		if (err != nil) {
			log.Println(err.Error())
		}

		var scheme string = "http://";
		if (r.TLS != nil) {
			scheme = "https://"
		}

		textures.Cape = &data.Cape{
			Url: scheme + r.Host + capeUrl.String(),
			Hash: capeRec.CalculateHash(),
		}
	}

	response,_ := json.Marshal(textures)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
