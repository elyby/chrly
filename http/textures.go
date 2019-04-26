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
	skin, err := cfg.SkinsRepo.FindByUsername(username)
	if err == nil && skin.SkinId != 0 {
		textures = &mojang.TexturesResponse{
			Skin: &mojang.SkinTexturesResponse{
				Url: skin.Url,
			},
		}

		if skin.IsSlim {
			textures.Skin.Metadata = &mojang.SkinTexturesMetadata{
				Model: "slim",
			}
		}

		_, err = cfg.CapesRepo.FindByUsername(username)
		if err == nil {
			var scheme = "http://"
			if request.TLS != nil {
				scheme = "https://"
			}

			textures.Cape = &mojang.CapeTexturesResponse{
				Url: scheme + request.Host + "/cloaks/" + username,
			}
		}
	} else {
		mojangTextures := <-cfg.MojangTexturesQueue.GetTexturesForUsername(username)
		if mojangTextures == nil {
			// TODO: test compatibility with exists authlibs
			response.WriteHeader(http.StatusNoContent)
			return
		}

		texturesProp := mojangTextures.DecodeTextures()
		if texturesProp == nil {
			// TODO: test compatibility with exists authlibs
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
