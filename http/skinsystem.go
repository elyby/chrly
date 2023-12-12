package http

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"github.com/elyby/chrly/utils"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/model"
)

var timeNow = time.Now

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

type TexturesSigner interface {
	SignTextures(textures string) (string, error)
	GetPublicKey() (*rsa.PublicKey, error)
}

type Skinsystem struct {
	Emitter
	SkinsRepo               SkinsRepository
	CapesRepo               CapesRepository
	MojangTexturesProvider  MojangTexturesProvider
	TexturesSigner          TexturesSigner
	TexturesExtraParamName  string
	TexturesExtraParamValue string
}

type profile struct {
	Id              string
	Username        string
	Textures        *mojang.TexturesResponse
	CapeFile        io.Reader
	MojangTextures  string
	MojangSignature string
}

func (ctx *Skinsystem) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/skins/{username}", ctx.skinHandler).Methods(http.MethodGet)
	router.HandleFunc("/cloaks/{username}", ctx.capeHandler).Methods(http.MethodGet).Name("cloaks")
	router.HandleFunc("/textures/{username}", ctx.texturesHandler).Methods(http.MethodGet)
	router.HandleFunc("/textures/signed/{username}", ctx.signedTexturesHandler).Methods(http.MethodGet)
	router.HandleFunc("/profile/{username}", ctx.profileHandler).Methods(http.MethodGet)
	// Legacy
	router.HandleFunc("/skins", ctx.skinGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/cloaks", ctx.capeGetHandler).Methods(http.MethodGet)
	// Utils
	router.HandleFunc("/signature-verification-key.der", ctx.signatureVerificationKeyHandler).Methods(http.MethodGet)
	router.HandleFunc("/signature-verification-key.pem", ctx.signatureVerificationKeyHandler).Methods(http.MethodGet)

	return router
}

func (ctx *Skinsystem) skinHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := ctx.getProfile(request, true)
	if err != nil {
		panic(err)
	}

	if profile == nil || profile.Textures == nil || profile.Textures.Skin == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(response, request, profile.Textures.Skin.Url, 301)
}

func (ctx *Skinsystem) skinGetHandler(response http.ResponseWriter, request *http.Request) {
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username

	ctx.skinHandler(response, request)
}

func (ctx *Skinsystem) capeHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := ctx.getProfile(request, true)
	if err != nil {
		panic(err)
	}

	if profile == nil || profile.Textures == nil || (profile.CapeFile == nil && profile.Textures.Cape == nil) {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	if profile.CapeFile == nil {
		http.Redirect(response, request, profile.Textures.Cape.Url, 301)
	} else {
		request.Header.Set("Content-Type", "image/png")
		_, _ = io.Copy(response, profile.CapeFile)
	}
}

func (ctx *Skinsystem) capeGetHandler(response http.ResponseWriter, request *http.Request) {
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username

	ctx.capeHandler(response, request)
}

func (ctx *Skinsystem) texturesHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := ctx.getProfile(request, true)
	if err != nil {
		panic(err)
	}

	if profile == nil || profile.Textures == nil || (profile.Textures.Skin == nil && profile.Textures.Cape == nil) {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	responseData, _ := json.Marshal(profile.Textures)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseData)
}

func (ctx *Skinsystem) signedTexturesHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := ctx.getProfile(request, request.URL.Query().Get("proxy") != "")
	if err != nil {
		panic(err)
	}

	if profile == nil || profile.MojangTextures == "" {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	profileResponse := &mojang.SignedTexturesResponse{
		Id:   profile.Id,
		Name: profile.Username,
		Props: []*mojang.Property{
			{
				Name:      "textures",
				Signature: profile.MojangSignature,
				Value:     profile.MojangTextures,
			},
			{
				Name:  ctx.TexturesExtraParamName,
				Value: ctx.TexturesExtraParamValue,
			},
		},
	}

	responseJson, _ := json.Marshal(profileResponse)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseJson)
}

func (ctx *Skinsystem) profileHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := ctx.getProfile(request, true)
	if err != nil {
		panic(err)
	}

	if profile == nil {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	texturesPropContent := &mojang.TexturesProp{
		Timestamp:   utils.UnixMillisecond(timeNow()),
		ProfileID:   profile.Id,
		ProfileName: profile.Username,
		Textures:    profile.Textures,
	}

	texturesPropValueJson, _ := json.Marshal(texturesPropContent)
	texturesPropEncodedValue := base64.StdEncoding.EncodeToString(texturesPropValueJson)

	texturesProp := &mojang.Property{
		Name:  "textures",
		Value: texturesPropEncodedValue,
	}

	if request.URL.Query().Get("unsigned") == "false" {
		signature, err := ctx.TexturesSigner.SignTextures(texturesProp.Value)
		if err != nil {
			panic(err)
		}

		texturesProp.Signature = signature
	}

	profileResponse := &mojang.SignedTexturesResponse{
		Id:   profile.Id,
		Name: profile.Username,
		Props: []*mojang.Property{
			texturesProp,
			{
				Name:  ctx.TexturesExtraParamName,
				Value: ctx.TexturesExtraParamValue,
			},
		},
	}

	responseJson, _ := json.Marshal(profileResponse)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseJson)
}

func (ctx *Skinsystem) signatureVerificationKeyHandler(response http.ResponseWriter, request *http.Request) {
	publicKey, err := ctx.TexturesSigner.GetPublicKey()
	if err != nil {
		panic(err)
	}

	asn1Bytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}

	if strings.HasSuffix(request.URL.Path, ".pem") {
		publicKeyBlock := pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: asn1Bytes,
		}

		publicKeyPemBytes := pem.EncodeToMemory(&publicKeyBlock)

		response.Header().Set("Content-Disposition", "attachment; filename=\"yggdrasil_session_pubkey.pem\"")
		_, _ = response.Write(publicKeyPemBytes)
	} else {
		response.Header().Set("Content-Type", "application/octet-stream")
		response.Header().Set("Content-Disposition", "attachment; filename=\"yggdrasil_session_pubkey.der\"")
		_, _ = response.Write(asn1Bytes)
	}
}

// TODO: in v5 should be extracted into some ProfileProvider interface,
//
//	which will encapsulate all logics, declared in this method
func (ctx *Skinsystem) getProfile(request *http.Request, proxy bool) (*profile, error) {
	username := parseUsername(mux.Vars(request)["username"])

	skin, err := ctx.SkinsRepo.FindSkinByUsername(username)
	if err != nil {
		return nil, err
	}

	profile := &profile{
		Id:              "",
		Username:        "",
		Textures:        &mojang.TexturesResponse{}, // Field must be initialized to avoid "null" after json encoding
		CapeFile:        nil,
		MojangTextures:  "",
		MojangSignature: "",
	}

	if skin != nil {
		profile.Id = strings.Replace(skin.Uuid, "-", "", -1)
		profile.Username = skin.Username
	}

	if skin != nil && skin.SkinId != 0 {
		profile.Textures.Skin = &mojang.SkinTexturesResponse{
			Url: skin.Url,
		}

		if skin.IsSlim {
			profile.Textures.Skin.Metadata = &mojang.SkinTexturesMetadata{
				Model: "slim",
			}
		}

		cape, _ := ctx.CapesRepo.FindCapeByUsername(username)
		if cape != nil {
			profile.CapeFile = cape.File
			profile.Textures.Cape = &mojang.CapeTexturesResponse{
				// Use statically http since the application doesn't support TLS
				Url: "http://" + request.Host + "/cloaks/" + username,
			}
		}

		profile.MojangTextures = skin.MojangTextures
		profile.MojangSignature = skin.MojangSignature
	} else if proxy {
		mojangProfile, err := ctx.MojangTexturesProvider.GetForUsername(username)
		// If we at least know something about a user,
		// than we can ignore an error and return profile without textures
		if err != nil && profile.Id != "" {
			return profile, nil
		}

		if err != nil || mojangProfile == nil {
			return nil, err
		}

		decodedTextures, err := mojangProfile.DecodeTextures()
		if err != nil {
			return nil, err
		}

		// There might be no textures property
		if decodedTextures != nil {
			profile.Textures = decodedTextures.Textures
		}

		var texturesProp *mojang.Property
		for _, prop := range mojangProfile.Props {
			if prop.Name == "textures" {
				texturesProp = prop
				break
			}
		}

		if texturesProp != nil {
			profile.MojangTextures = texturesProp.Value
			profile.MojangSignature = texturesProp.Signature
		}

		// If user id is unknown at this point, then use values from Mojang profile
		if profile.Id == "" {
			profile.Id = mojangProfile.Id
			profile.Username = mojangProfile.Name
		}
	} else {
		return nil, nil
	}

	return profile, nil
}

func parseUsername(username string) string {
	return strings.TrimSuffix(username, ".png")
}
