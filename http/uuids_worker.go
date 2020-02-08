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
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = http.HandlerFunc(NotFound)

	router.Handle("/api/worker/mojang-uuid/{username}", http.HandlerFunc(ctx.GetUUID)).Methods("GET")

	return router
}

func (ctx *UUIDsWorker) GetUUID(response http.ResponseWriter, request *http.Request) {
	username := parseUsername(mux.Vars(request)["username"])
	profile, err := ctx.UUIDsProvider.GetUuid(username)
	if err != nil {
		ctx.Emit("uuids_provider:error", err) // TODO: do I need emitter here?
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
