package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"ely.by/chrly/internal/db"
	"ely.by/chrly/internal/profiles"
)

type ProfilesManager interface {
	PersistProfile(ctx context.Context, profile *db.Profile) error
	RemoveProfileByUuid(ctx context.Context, uuid string) error
}

type Api struct {
	ProfilesManager
}

func (ctx *Api) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/profiles", ctx.postProfileHandler).Methods(http.MethodPost)
	router.HandleFunc("/profiles/{uuid}", ctx.deleteProfileByUuidHandler).Methods(http.MethodDelete)

	return router
}

func (ctx *Api) postProfileHandler(resp http.ResponseWriter, req *http.Request) {
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

	err = ctx.PersistProfile(req.Context(), profile)
	if err != nil {
		var v *profiles.ValidationError
		if errors.As(err, &v) {
			apiBadRequest(resp, v.Errors)
			return
		}

		apiServerError(resp, fmt.Errorf("unable to save profile to db: %w", err))
		return
	}

	resp.WriteHeader(http.StatusCreated)
}

func (ctx *Api) deleteProfileByUuidHandler(resp http.ResponseWriter, req *http.Request) {
	uuid := mux.Vars(req)["uuid"]
	err := ctx.ProfilesManager.RemoveProfileByUuid(req.Context(), uuid)
	if err != nil {
		apiServerError(resp, fmt.Errorf("unable to delete profile from db: %w", err))
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}
