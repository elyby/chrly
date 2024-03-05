package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"ely.by/chrly/internal/db"
	"ely.by/chrly/internal/mojang"
	"ely.by/chrly/internal/utils"
)

var timeNow = time.Now

type ProfilesProvider interface {
	FindProfileByUsername(ctx context.Context, username string, allowProxy bool) (*db.Profile, error)
}

// SignerService uses context because in the future we may separate this logic as an external microservice
type SignerService interface {
	Sign(ctx context.Context, data string) (string, error)
	GetPublicKey(ctx context.Context, format string) (string, error)
}

type Skinsystem struct {
	ProfilesProvider
	SignerService
	TexturesExtraParamName  string
	TexturesExtraParamValue string
}

func (s *Skinsystem) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/skins/{username}", s.skinHandler).Methods(http.MethodGet)
	router.HandleFunc("/cloaks/{username}", s.capeHandler).Methods(http.MethodGet)
	// TODO: alias /capes/{username}?
	router.HandleFunc("/textures/{username}", s.texturesHandler).Methods(http.MethodGet)
	router.HandleFunc("/textures/signed/{username}", s.signedTexturesHandler).Methods(http.MethodGet)
	router.HandleFunc("/profile/{username}", s.profileHandler).Methods(http.MethodGet)
	// Legacy
	router.HandleFunc("/skins", s.skinGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/cloaks", s.capeGetHandler).Methods(http.MethodGet)
	// Utils
	router.HandleFunc("/signature-verification-key.{format:(?:pem|der)}", s.signatureVerificationKeyHandler).Methods(http.MethodGet)

	return router
}

func (s *Skinsystem) skinHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := s.ProfilesProvider.FindProfileByUsername(request.Context(), parseUsername(mux.Vars(request)["username"]), true)
	if err != nil {
		apiServerError(response, fmt.Errorf("unable to retrieve a profile: %w", err))
		return
	}

	if profile == nil || profile.SkinUrl == "" {
		response.WriteHeader(http.StatusNotFound)
	}

	http.Redirect(response, request, profile.SkinUrl, http.StatusMovedPermanently)
}

func (s *Skinsystem) skinGetHandler(response http.ResponseWriter, request *http.Request) {
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username

	s.skinHandler(response, request)
}

func (s *Skinsystem) capeHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := s.ProfilesProvider.FindProfileByUsername(request.Context(), parseUsername(mux.Vars(request)["username"]), true)
	if err != nil {
		apiServerError(response, fmt.Errorf("unable to retrieve a profile: %w", err))
		return
	}

	if profile == nil || profile.CapeUrl == "" {
		response.WriteHeader(http.StatusNotFound)
	}

	http.Redirect(response, request, profile.CapeUrl, http.StatusMovedPermanently)
}

func (s *Skinsystem) capeGetHandler(response http.ResponseWriter, request *http.Request) {
	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username

	s.capeHandler(response, request)
}

func (s *Skinsystem) texturesHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := s.ProfilesProvider.FindProfileByUsername(request.Context(), mux.Vars(request)["username"], true)
	if err != nil {
		apiServerError(response, fmt.Errorf("unable to retrieve a profile: %w", err))
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

func (s *Skinsystem) signedTexturesHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := s.ProfilesProvider.FindProfileByUsername(
		request.Context(),
		mux.Vars(request)["username"],
		getToBool(request.URL.Query().Get("proxy")),
	)
	if err != nil {
		apiServerError(response, fmt.Errorf("unable to retrieve a profile: %w", err))
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
				Name:  s.TexturesExtraParamName,
				Value: s.TexturesExtraParamValue,
			},
		},
	}

	responseJson, _ := json.Marshal(profileResponse)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseJson)
}

func (s *Skinsystem) profileHandler(response http.ResponseWriter, request *http.Request) {
	profile, err := s.ProfilesProvider.FindProfileByUsername(request.Context(), mux.Vars(request)["username"], true)
	if err != nil {
		apiServerError(response, fmt.Errorf("unable to retrieve a profile: %w", err))
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
		signature, err := s.SignerService.Sign(request.Context(), texturesProp.Value)
		if err != nil {
			apiServerError(response, fmt.Errorf("unable to sign textures: %w", err))
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
				Name:  s.TexturesExtraParamName,
				Value: s.TexturesExtraParamValue,
			},
		},
	}

	responseJson, _ := json.Marshal(profileResponse)
	response.Header().Set("Content-Type", "application/json")
	_, _ = response.Write(responseJson)
}

func (s *Skinsystem) signatureVerificationKeyHandler(response http.ResponseWriter, request *http.Request) {
	format := mux.Vars(request)["format"]
	publicKey, err := s.SignerService.GetPublicKey(request.Context(), format)
	if err != nil {
		apiServerError(response, fmt.Errorf("unable to retrieve public key: %w", err))
		return
	}

	if format == "pem" {
		response.Header().Set("Content-Type", "application/x-pem-file")
		response.Header().Set("Content-Disposition", `attachment; filename="yggdrasil_session_pubkey.pem"`)
	} else {
		response.Header().Set("Content-Type", "application/octet-stream")
		response.Header().Set("Content-Disposition", `attachment; filename="yggdrasil_session_pubkey.der"`)
	}

	_, _ = io.WriteString(response, publicKey)
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
