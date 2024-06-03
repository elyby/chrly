package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/huandu/xstrings"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"

	"ely.by/chrly/internal/db"
	"ely.by/chrly/internal/otel"
	"ely.by/chrly/internal/profiles"
)

type ProfilesManager interface {
	PersistProfile(ctx context.Context, profile *db.Profile) error
	RemoveProfileByUuid(ctx context.Context, uuid string) error
}

func NewProfilesApi(profilesManager ProfilesManager) (*ProfilesApi, error) {
	metrics, err := newProfilesApiMetrics(otel.GetMeter())
	if err != nil {
		return nil, err
	}

	return &ProfilesApi{
		ProfilesManager: profilesManager,
		metrics:         metrics,
	}, nil
}

type ProfilesApi struct {
	ProfilesManager

	metrics *profilesApiMetrics
}

func (p *ProfilesApi) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", p.postProfileHandler).Methods(http.MethodPost)
	router.HandleFunc("/{uuid}", p.deleteProfileByUuidHandler).Methods(http.MethodDelete)

	return router
}

func (p *ProfilesApi) postProfileHandler(resp http.ResponseWriter, req *http.Request) {
	p.metrics.UploadProfileRequest.Add(req.Context(), 1)

	err := req.ParseForm()
	if err != nil {
		apiBadRequest(resp, map[string][]string{
			"body": {"The body of the request must be a valid url-encoded string"},
		})
		return
	}

	profile := &db.Profile{
		Uuid:            req.Form.Get("uuid"),
		Username:        req.Form.Get("username"),
		SkinUrl:         req.Form.Get("skinUrl"),
		SkinModel:       req.Form.Get("skinModel"),
		CapeUrl:         req.Form.Get("capeUrl"),
		MojangTextures:  req.Form.Get("mojangTextures"),
		MojangSignature: req.Form.Get("mojangSignature"),
	}

	err = p.PersistProfile(req.Context(), profile)
	if err != nil {
		var v *profiles.ValidationError
		if errors.As(err, &v) {
			// Manager returns ValidationError according to the struct fields names.
			// They are uppercased, but otherwise the same as the names in the API.
			// So to make them consistent it's enough just to make the first lowercased.
			for field, errors := range v.Errors {
				v.Errors[xstrings.FirstRuneToLower(field)] = errors
				delete(v.Errors, field)
			}

			apiBadRequest(resp, v.Errors)

			return
		}

		apiServerError(resp, req, fmt.Errorf("unable to save profile to db: %w", err))
		return
	}

	resp.WriteHeader(http.StatusCreated)
}

func (p *ProfilesApi) deleteProfileByUuidHandler(resp http.ResponseWriter, req *http.Request) {
	p.metrics.DeleteProfileRequest.Add(req.Context(), 1)

	uuid := mux.Vars(req)["uuid"]
	err := p.ProfilesManager.RemoveProfileByUuid(req.Context(), uuid)
	if err != nil {
		apiServerError(resp, req, fmt.Errorf("unable to delete profile from db: %w", err))
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

func newProfilesApiMetrics(meter metric.Meter) (*profilesApiMetrics, error) {
	m := &profilesApiMetrics{}
	var errors, err error

	m.UploadProfileRequest, err = meter.Int64Counter("chrly.app.profiles.upload.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	m.DeleteProfileRequest, err = meter.Int64Counter("chrly.app.profiles.delete.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	return m, errors
}

type profilesApiMetrics struct {
	UploadProfileRequest metric.Int64Counter
	DeleteProfileRequest metric.Int64Counter
}
