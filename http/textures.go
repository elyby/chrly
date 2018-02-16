package http

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/elyby/chrly/model"
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

func (cfg *Config) Textures(response http.ResponseWriter, request *http.Request) {
	cfg.Logger.IncCounter("textures.request", 1)
	username := parseUsername(mux.Vars(request)["username"])

	skin, err := cfg.SkinsRepo.FindByUsername(username)
	if err != nil || skin.SkinId == 0 {
		if skin == nil {
			skin = &model.Skin{}
		}

		skin.Url = "http://skins.minecraft.net/MinecraftSkins/" + username + ".png"
		skin.Hash = string(buildNonElyTexturesHash(username))
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

	cape, err := cfg.CapesRepo.FindByUsername(username)
	if err == nil {
		var scheme string = "http://"
		if request.TLS != nil {
			scheme = "https://"
		}

		textures.Cape = &Cape{
			Url:  scheme + request.Host + "/cloaks/" + username,
			Hash: calculateCapeHash(cape),
		}
	}

	responseData, _ := json.Marshal(textures)
	response.Header().Set("Content-Type", "application/json")
	response.Write(responseData)
}

func calculateCapeHash(cape *model.Cape) string {
	hasher := md5.New()
	io.Copy(hasher, cape.File)

	return hex.EncodeToString(hasher.Sum(nil))
}

func buildNonElyTexturesHash(username string) string {
	hour := getCurrentHour()
	hasher := md5.New()
	hasher.Write([]byte("non-ely-" + strconv.FormatInt(hour, 10) + "-" + username))

	return hex.EncodeToString(hasher.Sum(nil))
}

var timeNow = time.Now

func getCurrentHour() int64 {
	n := timeNow()
	return time.Date(n.Year(), n.Month(), n.Day(), n.Hour(), 0, 0, 0, time.UTC).Unix()
}
