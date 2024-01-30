package http

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/elyby/chrly/db"
	"github.com/elyby/chrly/mojang"
	"github.com/elyby/chrly/utils"
)

var timeNow = time.Now

type ProfilesProvider interface {
	FindProfileByUsername(username string, allowProxy bool) (*db.Profile, error)
}

type TexturesSigner interface {
	SignTextures(textures string) (string, error)
	GetPublicKey() (*rsa.PublicKey, error)
}

type Skinsystem struct {
	ProfilesProvider
	TexturesSigner
	TexturesExtraParamName  string
	TexturesExtraParamValue string
}

func (ctx *Skinsystem) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/skins/{username}", ctx.skinHandler).Methods(http.MethodGet)
	router.HandleFunc("/cloaks/{username}", ctx.capeHandler).Methods(http.MethodGet)
	// TODO: alias /capes/{username}?
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
	profile, err := ctx.ProfilesProvider.FindProfileByUsername(parseUsername(mux.Vars(request)["username"]), true)
	if err != nil {
		apiServerError(response, "Unable to retrieve a skin", err)
		return
	}

	if profile == nil || profile.SkinUrl == "" {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(response, request, profile.SkinUrl, http.StatusMovedPermanently)
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
	profile, err := ctx.ProfilesProvider.FindProfileByUsername(parseUsername(mux.Vars(request)["username"]), true)
	if err != nil {
		apiServerError(response, "Unable to retrieve a cape", err)
		return
	}

	if profile == nil || profile.CapeUrl == "" {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(response, request, profile.CapeUrl, http.StatusMovedPermanently)
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
	profile, err := ctx.ProfilesProvider.FindProfileByUsername(mux.Vars(request)["username"], true)
	if err != nil {
		apiServerError(response, "Unable to retrieve a profile", err)
		return
	}

	if profile == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	if profile.SkinUrl == "" && profile.CapeUrl == "" {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	textures := texturesFromProfile(profile)

	responseData, _ := json.Marshal(textures)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseData)
}

func (ctx *Skinsystem) signedTexturesHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := ctx.ProfilesProvider.FindProfileByUsername(
		mux.Vars(request)["username"],
		getToBool(request.URL.Query().Get("proxy")),
	)
	if err != nil {
		apiServerError(response, "Unable to retrieve a profile", err)
		return
	}

	if profile == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	if profile.MojangTextures == "" {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	profileResponse := &mojang.ProfileResponse{
		Id:   profile.Uuid,
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
	profile, err := ctx.ProfilesProvider.FindProfileByUsername(mux.Vars(request)["username"], true)
	if err != nil {
		apiServerError(response, "Unable to retrieve a profile", err)
		return
	}

	if profile == nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	texturesPropContent := &mojang.TexturesProp{
		Timestamp:   utils.UnixMillisecond(timeNow()),
		ProfileID:   profile.Uuid,
		ProfileName: profile.Username,
		Textures:    texturesFromProfile(profile),
	}

	texturesPropValueJson, _ := json.Marshal(texturesPropContent)
	texturesPropEncodedValue := base64.StdEncoding.EncodeToString(texturesPropValueJson)

	texturesProp := &mojang.Property{
		Name:  "textures",
		Value: texturesPropEncodedValue,
	}

	if request.URL.Query().Has("unsigned") && !getToBool(request.URL.Query().Get("unsigned")) {
		signature, err := ctx.TexturesSigner.SignTextures(texturesProp.Value)
		if err != nil {
			apiServerError(response, "Unable to sign textures", err)
			return
		}

		texturesProp.Signature = signature
	}

	profileResponse := &mojang.ProfileResponse{
		Id:   profile.Uuid,
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

func parseUsername(username string) string {
	return strings.TrimSuffix(username, ".png")
}

func getToBool(v string) bool {
	return v == "true" || v == "1" || v == "yes"
}

func texturesFromProfile(profile *db.Profile) *mojang.TexturesResponse {
	var skin *mojang.SkinTexturesResponse
	if profile.SkinUrl != "" {
		skin = &mojang.SkinTexturesResponse{
			Url: profile.SkinUrl,
		}
		if profile.SkinModel != "" {
			skin.Metadata = &mojang.SkinTexturesMetadata{
				Model: profile.SkinModel,
			}
		}
	}

	var cape *mojang.CapeTexturesResponse
	if profile.CapeUrl != "" {
		cape = &mojang.CapeTexturesResponse{
			Url: profile.CapeUrl,
		}
	}

	return &mojang.TexturesResponse{
		Skin: skin,
		Cape: cape,
	}
}
