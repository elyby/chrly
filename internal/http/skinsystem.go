package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"

	"ely.by/chrly/internal/db"
	"ely.by/chrly/internal/mojang"
	"ely.by/chrly/internal/otel"
)

type ProfilesProvider interface {
	FindProfileByUsername(ctx context.Context, username string, allowProxy bool) (*db.Profile, error)
}

func NewSkinsystemApi(
	profilesProvider ProfilesProvider,
	texturesExtraParamName string,
	texturesExtraParamValue string,
) (*Skinsystem, error) {
	metrics, err := newSkinsystemMetrics(otel.GetMeter())
	if err != nil {
		return nil, err
	}

	return &Skinsystem{
		ProfilesProvider:        profilesProvider,
		TexturesExtraParamName:  texturesExtraParamName,
		TexturesExtraParamValue: texturesExtraParamValue,
		metrics:                 metrics,
	}, nil
}

type Skinsystem struct {
	ProfilesProvider
	TexturesExtraParamName  string
	TexturesExtraParamValue string
	metrics                 *skinsystemApiMetrics
}

func (s *Skinsystem) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/skins/{username}", s.skinHandler).Methods(http.MethodGet)
	router.HandleFunc("/cloaks/{username}", s.capeHandler).Methods(http.MethodGet)
	// TODO: alias /capes/{username}?
	router.HandleFunc("/textures/{username}", s.texturesHandler).Methods(http.MethodGet)
	router.HandleFunc("/textures/signed/{username}", s.signedTexturesHandler).Methods(http.MethodGet)
	// Legacy
	router.HandleFunc("/skins", s.legacySkinHandler).Methods(http.MethodGet)
	router.HandleFunc("/cloaks", s.legacyCapeHandler).Methods(http.MethodGet)

	return router
}

func (s *Skinsystem) skinHandler(response http.ResponseWriter, request *http.Request) {
	s.metrics.SkinRequest.Add(request.Context(), 1)

	s.skinHandlerWithUsername(response, request, mux.Vars(request)["username"])
}

func (s *Skinsystem) legacySkinHandler(response http.ResponseWriter, request *http.Request) {
	s.metrics.LegacySkinRequest.Add(request.Context(), 1)

	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	s.skinHandlerWithUsername(response, request, username)
}

func (s *Skinsystem) skinHandlerWithUsername(resp http.ResponseWriter, req *http.Request, username string) {
	profile, err := s.ProfilesProvider.FindProfileByUsername(req.Context(), parseUsername(username), true)
	if err != nil {
		apiServerError(resp, req, fmt.Errorf("unable to retrieve a profile: %w", err))
		return
	}

	if profile == nil || profile.SkinUrl == "" {
		resp.WriteHeader(http.StatusNotFound)
	}

	http.Redirect(resp, req, profile.SkinUrl, http.StatusMovedPermanently)
}

func (s *Skinsystem) capeHandler(response http.ResponseWriter, request *http.Request) {
	s.metrics.CapeRequest.Add(request.Context(), 1)

	s.capeHandlerWithUsername(response, request, mux.Vars(request)["username"])
}

func (s *Skinsystem) legacyCapeHandler(response http.ResponseWriter, request *http.Request) {
	s.metrics.CapeRequest.Add(request.Context(), 1)

	username := request.URL.Query().Get("name")
	if username == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	s.capeHandlerWithUsername(response, request, username)
}

func (s *Skinsystem) capeHandlerWithUsername(resp http.ResponseWriter, req *http.Request, username string) {
	profile, err := s.ProfilesProvider.FindProfileByUsername(req.Context(), parseUsername(username), true)
	if err != nil {
		apiServerError(resp, req, fmt.Errorf("unable to retrieve a profile: %w", err))
		return
	}

	if profile == nil || profile.CapeUrl == "" {
		resp.WriteHeader(http.StatusNotFound)
	}

	http.Redirect(resp, req, profile.CapeUrl, http.StatusMovedPermanently)
}

func (s *Skinsystem) texturesHandler(resp http.ResponseWriter, req *http.Request) {
	s.metrics.TexturesRequest.Add(req.Context(), 1)

	profile, err := s.ProfilesProvider.FindProfileByUsername(req.Context(), mux.Vars(req)["username"], true)
	if err != nil {
		apiServerError(resp, req, fmt.Errorf("unable to retrieve a profile: %w", err))
		return
	}

	if profile == nil {
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	if profile.SkinUrl == "" && profile.CapeUrl == "" {
		resp.WriteHeader(http.StatusNoContent)
		return
	}

	textures := texturesFromProfile(profile)

	responseData, _ := json.Marshal(textures)
	resp.Header().Set("Content-Type", "application/json")
	_, _ = resp.Write(responseData)
}

func (s *Skinsystem) signedTexturesHandler(resp http.ResponseWriter, req *http.Request) {
	s.metrics.SignedTexturesRequest.Add(req.Context(), 1)

	profile, err := s.ProfilesProvider.FindProfileByUsername(
		req.Context(),
		mux.Vars(req)["username"],
		getToBool(req.URL.Query().Get("proxy")),
	)
	if err != nil {
		apiServerError(resp, req, fmt.Errorf("unable to retrieve a profile: %w", err))
		return
	}

	if profile == nil {
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	if profile.MojangTextures == "" {
		resp.WriteHeader(http.StatusNoContent)
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
	resp.Header().Set("Content-Type", "application/json")
	_, _ = resp.Write(responseJson)
}

func parseUsername(username string) string {
	return strings.TrimSuffix(username, ".png")
}

func getToBool(v string) bool {
	return v == "1" || v == "true" || v == "yes"
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

func newSkinsystemMetrics(meter metric.Meter) (*skinsystemApiMetrics, error) {
	m := &skinsystemApiMetrics{}
	var errors, err error

	m.SkinRequest, err = meter.Int64Counter("chrly.app.skinsystem.skin.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	m.LegacySkinRequest, err = meter.Int64Counter("chrly.app.skinsystem.legacy_skin.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	m.CapeRequest, err = meter.Int64Counter("chrly.app.skinsystem.cape.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	m.LegacyCapeRequest, err = meter.Int64Counter("chrly.app.skinsystem.legacy_cape.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	m.TexturesRequest, err = meter.Int64Counter("chrly.app.skinsystem.textures.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	m.SignedTexturesRequest, err = meter.Int64Counter("chrly.app.skinsystem.signed_textures.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	return m, errors
}

type skinsystemApiMetrics struct {
	SkinRequest           metric.Int64Counter
	LegacySkinRequest     metric.Int64Counter
	CapeRequest           metric.Int64Counter
	LegacyCapeRequest     metric.Int64Counter
	TexturesRequest       metric.Int64Counter
	SignedTexturesRequest metric.Int64Counter
}
