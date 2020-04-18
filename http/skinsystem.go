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
	FindByUsername(username string) (*model.Skin, error)
	FindByUserId(id int) (*model.Skin, error)
	Save(skin *model.Skin) error
	RemoveByUserId(id int) error
	RemoveByUsername(username string) error
}

type CapesRepository interface {
	FindByUsername(username string) (*model.Cape, error)
}

// TODO: can I get rid of this?
type SkinNotFoundError struct {
	Who string
}

func (e SkinNotFoundError) Error() string {
	return "skin data not found"
}

type CapeNotFoundError struct {
	Who string
}

// TODO: can I get rid of this?
func (e CapeNotFoundError) Error() string {
	return "cape file not found"
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
	rec, err := ctx.SkinsRepo.FindByUsername(username)
	if err == nil && rec.SkinId != 0 {
		http.Redirect(response, request, rec.Url, 301)
		return
	}

	mojangTextures, err := ctx.MojangTexturesProvider.GetForUsername(username)
	if err != nil || mojangTextures == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	texturesProp := mojangTextures.DecodeTextures()
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
	rec, err := ctx.CapesRepo.FindByUsername(username)
	if err == nil {
		request.Header.Set("Content-Type", "image/png")
		_, _ = io.Copy(response, rec.File)
		return
	}

	mojangTextures, err := ctx.MojangTexturesProvider.GetForUsername(username)
	if err != nil || mojangTextures == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	texturesProp := mojangTextures.DecodeTextures()
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
	skin, skinErr := ctx.SkinsRepo.FindByUsername(username)
	_, capeErr := ctx.CapesRepo.FindByUsername(username)
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
		mojangTextures, err := ctx.MojangTexturesProvider.GetForUsername(username)
		if err != nil || mojangTextures == nil {
			response.WriteHeader(http.StatusNoContent)
			return
		}

		texturesProp := mojangTextures.DecodeTextures()
		if texturesProp == nil {
			ctx.Emit("skinsystem:error", errors.New("unable to find textures property"))
			apiServerError(response)
			return
		}

		textures = texturesProp.Textures
		// TODO: return 204 in case when there is no skin and cape on mojang textures
	}

	responseData, _ := json.Marshal(textures)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseData)
}

func (ctx *Skinsystem) signedTexturesHandler(response http.ResponseWriter, request *http.Request) {
	username := parseUsername(mux.Vars(request)["username"])

	var responseData *mojang.SignedTexturesResponse

	rec, err := ctx.SkinsRepo.FindByUsername(username)
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
		Name:  getStringOrDefault(ctx.TexturesExtraParamName, "chrly"),
		Value: getStringOrDefault(ctx.TexturesExtraParamValue, "how do you tame a horse in Minecraft?"),
	})

	responseJson, _ := json.Marshal(responseData)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseJson)
}

func parseUsername(username string) string {
	return strings.TrimSuffix(username, ".png")
}

func getStringOrDefault(value string, def string) string {
	if value != "" {
		return value
	}

	return def
}
