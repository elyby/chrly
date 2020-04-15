package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/mojangtextures"
)

type UuidsProvider interface {
	GetUuid(username string) (*mojang.ProfileInfo, error)
}

type UUIDsWorker struct {
	Emitter
	UUIDsProvider mojangtextures.UUIDsProvider
}

func (ctx *UUIDsWorker) CreateHandler() *mux.Router {
	requestEventsMiddleware := CreateRequestEventsMiddleware(ctx.Emitter, "skinsystem") // This prefix should be unified

	router := mux.NewRouter().StrictSlash(true)
	router.Use(requestEventsMiddleware)

	router.Handle("/api/worker/mojang-uuid/{username}", http.HandlerFunc(ctx.GetUUID)).Methods("GET")

	// 404
	// NotFoundHandler doesn't call for registered middlewares, so we must wrap it manually.
	// See https://github.com/gorilla/mux/issues/416#issuecomment-600079279
	router.NotFoundHandler = requestEventsMiddleware(http.HandlerFunc(NotFound))

	return router
}

func (ctx *UUIDsWorker) GetUUID(response http.ResponseWriter, request *http.Request) {
	username := parseUsername(mux.Vars(request)["username"])
	profile, err := ctx.UUIDsProvider.GetUuid(username)
	if err != nil {
		if _, ok := err.(*mojang.TooManyRequestsError); ok {
			response.WriteHeader(http.StatusTooManyRequests)
			return
		}

		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusInternalServerError)
		result, _ := json.Marshal(map[string]interface{}{
			"provider": err.Error(),
		})
		_, _ = response.Write(result)
		return
	}

	if profile == nil {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	responseData, _ := json.Marshal(profile)
	_, _ = response.Write(responseData)
}
