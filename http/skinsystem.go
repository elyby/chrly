package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/model"
)

type SkinsRepository interface {
	FindSkinByUsername(username string) (*model.Skin, error)
	FindSkinByUserId(id int) (*model.Skin, error)
	SaveSkin(skin *model.Skin) error
	RemoveSkinByUserId(id int) error
	RemoveSkinByUsername(username string) error
}

type CapesRepository interface {
	FindCapeByUsername(username string) (*model.Cape, error)
}

type MojangTexturesProvider interface {
	GetForUsername(username string) (*mojang.SignedTexturesResponse, error)
}

type Skinsystem struct {
	Emitter
	SkinsRepo               SkinsRepository
	CapesRepo               CapesRepository
	MojangTexturesProvider  MojangTexturesProvider
	TexturesExtraParamName  string
	TexturesExtraParamValue string
}

func (ctx *Skinsystem) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/skins/{username}", ctx.skinHandler).Methods(http.MethodGet)
	router.HandleFunc("/cloaks/{username}", ctx.capeHandler).Methods(http.MethodGet).Name("cloaks")
	router.HandleFunc("/textures/{username}", ctx.texturesHandler).Methods(http.MethodGet)
	router.HandleFunc("/textures/signed/{username}", ctx.signedTexturesHandler).Methods(http.MethodGet)
	// Legacy
	router.HandleFunc("/skins", ctx.skinGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/cloaks", ctx.capeGetHandler).Methods(http.MethodGet)

	return router
}

func (ctx *Skinsystem) skinHandler(response http.ResponseWriter, request *http.Request) {
	username := parseUsername(mux.Vars(request)["username"])
	rec, err := ctx.SkinsRepo.FindSkinByUsername(username)
	if err == nil && rec != nil && rec.SkinId != 0 {
		http.Redirect(response, request, rec.Url, 301)
		return
	}

	mojangTextures, err := ctx.MojangTexturesProvider.GetForUsername(username)
	if err != nil || mojangTextures == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	texturesProp, _ := mojangTextures.DecodeTextures()
	skin := texturesProp.Textures.Skin
	if skin == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(response, request, skin.Url, 301)
}

func (ctx *Skinsystem) skinGetHandler(response http.ResponseWriter, request *http.Request) {
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username
	mux.Vars(request)["converted"] = "1"

	ctx.skinHandler(response, request)
}

func (ctx *Skinsystem) capeHandler(response http.ResponseWriter, request *http.Request) {
	username := parseUsername(mux.Vars(request)["username"])
	rec, err := ctx.CapesRepo.FindCapeByUsername(username)
	if err == nil && rec != nil {
		request.Header.Set("Content-Type", "image/png")
		_, _ = io.Copy(response, rec.File)
		return
	}

	mojangTextures, err := ctx.MojangTexturesProvider.GetForUsername(username)
	if err != nil || mojangTextures == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	texturesProp, _ := mojangTextures.DecodeTextures()
	cape := texturesProp.Textures.Cape
	if cape == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(response, request, cape.Url, 301)
}

func (ctx *Skinsystem) capeGetHandler(response http.ResponseWriter, request *http.Request) {
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username
	mux.Vars(request)["converted"] = "1"

	ctx.capeHandler(response, request)
}

func (ctx *Skinsystem) texturesHandler(response http.ResponseWriter, request *http.Request) {
	username := parseUsername(mux.Vars(request)["username"])

	var textures *mojang.TexturesResponse
	skin, skinErr := ctx.SkinsRepo.FindSkinByUsername(username)
	cape, capeErr := ctx.CapesRepo.FindCapeByUsername(username)
	if (skinErr == nil && skin != nil && skin.SkinId != 0) || (capeErr == nil && cape != nil) {
		textures = &mojang.TexturesResponse{}
		if skinErr == nil && skin != nil && skin.SkinId != 0 {
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

		if capeErr == nil && cape != nil {
			textures.Cape = &mojang.CapeTexturesResponse{
				// Use statically http since the application doesn't support TLS
				Url: "http://" + request.Host + "/cloaks/" + username,
			}
		}
	} else {
		mojangTextures, err := ctx.MojangTexturesProvider.GetForUsername(username)
		if err != nil || mojangTextures == nil {
			response.WriteHeader(http.StatusNoContent)
			return
		}

		texturesProp, _ := mojangTextures.DecodeTextures()
		if texturesProp == nil {
			ctx.Emit("skinsystem:error", errors.New("unable to find textures property"))
			apiServerError(response)
			return
		}

		textures = texturesProp.Textures
		if textures.Skin == nil && textures.Cape == nil {
			response.WriteHeader(http.StatusNoContent)
			return
		}
	}

	responseData, _ := json.Marshal(textures)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseData)
}

func (ctx *Skinsystem) signedTexturesHandler(response http.ResponseWriter, request *http.Request) {
	username := parseUsername(mux.Vars(request)["username"])

	var responseData *mojang.SignedTexturesResponse

	rec, err := ctx.SkinsRepo.FindSkinByUsername(username)
	if err == nil && rec != nil && rec.SkinId != 0 && rec.MojangTextures != "" {
		responseData = &mojang.SignedTexturesResponse{
			Id:   strings.Replace(rec.Uuid, "-", "", -1),
			Name: rec.Username,
			Props: []*mojang.Property{
				{
					Name:      "textures",
					Signature: rec.MojangSignature,
					Value:     rec.MojangTextures,
				},
			},
		}
	} else if request.URL.Query().Get("proxy") != "" {
		mojangTextures, err := ctx.MojangTexturesProvider.GetForUsername(username)
		if err == nil && mojangTextures != nil {
			responseData = mojangTextures
		}
	}

	if responseData == nil {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	responseData.Props = append(responseData.Props, &mojang.Property{
		Name:  ctx.TexturesExtraParamName,
		Value: ctx.TexturesExtraParamValue,
	})

	responseJson, _ := json.Marshal(responseData)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseJson)
}

func parseUsername(username string) string {
	return strings.TrimSuffix(username, ".png")
}
