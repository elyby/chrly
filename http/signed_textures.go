package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/elyby/chrly/api/mojang"
)

func (cfg *Config) SignedTextures(response http.ResponseWriter, request *http.Request) {
	cfg.Logger.IncCounter("signed_textures.request", 1)
	username := parseUsername(mux.Vars(request)["username"])

	var responseData *mojang.SignedTexturesResponse

	rec, err := cfg.SkinsRepo.FindByUsername(username)
	if err == nil && rec.SkinId != 0 && rec.MojangTextures != "" {
		responseData = &mojang.SignedTexturesResponse{
			Id:   strings.Replace(rec.Uuid, "-", "", -1),
			Name: rec.Username,
			Props: []*mojang.Property{
				{
					Name:      "textures",
					Signature: rec.MojangSignature,
					Value:     rec.MojangTextures,
				},
				{
					Name:  "chrly",
					Value: "how do you tame a horse in Minecraft?",
				},
			},
		}
	} else if request.URL.Query().Get("proxy") != "" {
		responseData = <-cfg.MojangTexturesQueue.GetTexturesForUsername(username)
	}

	if responseData == nil {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	responseJson, _ := json.Marshal(responseData)
	response.Header().Set("Content-Type", "application/json")
	response.Write(responseJson)
}
