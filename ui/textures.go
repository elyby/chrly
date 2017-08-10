package ui

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"crypto/md5"
	"encoding/hex"
	"io"

	"elyby/minecraft-skinsystem/model"
	"elyby/minecraft-skinsystem/utils"
)

type texturesResponse struct {
	Skin *Skin `json:"SKIN"`
	Cape *Cape `json:"CAPE,omitempty"`
}

type Skin struct {
	Url      string        `json:"url"`
	Hash     string        `json:"hash"`
	Metadata *skinMetadata `json:"metadata,omitempty"`
}

type skinMetadata struct {
	Model string `json:"model"`
}

type Cape struct {
	Url  string `json:"url"`
	Hash string `json:"hash"`
}

func (s *uiService) Textures(response http.ResponseWriter, request *http.Request) {
	s.logger.IncCounter("textures.request", 1)
	username := utils.ParseUsername(mux.Vars(request)["username"])

	skin, err := s.skinsRepo.FindByUsername(username)
	if err != nil || skin.SkinId == 0 {
		skin.Url = "http://skins.minecraft.net/MinecraftSkins/" + username + ".png"
		skin.Hash = string(utils.BuildNonElyTexturesHash(username))
	} else {
		skin.Url = utils.BuildElyUrl(skin.Url)
	}

	textures := texturesResponse{
		Skin: &Skin{
			Url:  skin.Url,
			Hash: skin.Hash,
		},
	}

	if skin.IsSlim {
		textures.Skin.Metadata = &skinMetadata{
			Model: "slim",
		}
	}

	cape, err := s.capesRepo.FindByUsername(username)
	if err == nil {
		// TODO: восстановить функционал получения ссылки на плащ
		// capeUrl, err := services.Router.Get("cloaks").URL("username", username)
		capeUrl := "/capes/" + username
		if err != nil {
			s.logger.Error(err.Error())
		}

		var scheme string = "http://"
		if request.TLS != nil {
			scheme = "https://"
		}

		textures.Cape = &Cape{
			// Url:  scheme + request.Host + capeUrl.String(),
			Url:  scheme + request.Host + capeUrl,
			Hash: calculateCapeHash(cape),
		}
	}

	responseData,_ := json.Marshal(textures)
	response.Header().Set("Content-Type", "application/json")
	response.Write(responseData)
}

func calculateCapeHash(cape model.Cape) string {
	hasher := md5.New()
	io.Copy(hasher, cape.File)

	return hex.EncodeToString(hasher.Sum(nil))
}
