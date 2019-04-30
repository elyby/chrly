package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/elyby/chrly/api/mojang"
)

func (cfg *Config) Textures(response http.ResponseWriter, request *http.Request) {
	cfg.Logger.IncCounter("textures.request", 1)
	username := parseUsername(mux.Vars(request)["username"])

	var textures *mojang.TexturesResponse
	skin, skinErr := cfg.SkinsRepo.FindByUsername(username)
	_, capeErr := cfg.CapesRepo.FindByUsername(username)
	if (skinErr == nil && skin.SkinId != 0) || capeErr == nil {
		textures = &mojang.TexturesResponse{}

		if skinErr == nil && skin.SkinId != 0 {
			skinTextures := &mojang.SkinTexturesResponse{
				Url: skin.Url,
			}

			if skin.IsSlim {
				skinTextures.Metadata = &mojang.SkinTexturesMetadata{
					Model: "slim",
				}
			}

			textures.Skin = skinTextures
		}

		if capeErr == nil {
			textures.Cape = &mojang.CapeTexturesResponse{
				Url: request.URL.Scheme + "://" + request.Host + "/cloaks/" + username,
			}
		}
	} else {
		mojangTextures := <-cfg.MojangTexturesQueue.GetTexturesForUsername(username)
		if mojangTextures == nil {
			response.WriteHeader(http.StatusNoContent)
			return
		}

		texturesProp := mojangTextures.DecodeTextures()
		if texturesProp == nil {
			response.WriteHeader(http.StatusInternalServerError)
			cfg.Logger.Error("Unable to find textures property")
			return
		}

		textures = texturesProp.Textures
	}

	responseData, _ := json.Marshal(textures)
	response.Header().Set("Content-Type", "application/json")
	response.Write(responseData)
}
