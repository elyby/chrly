package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/elyby/chrly/api/mojang"
)

type MojangUuidsProvider interface {
	GetUuid(username string) (*mojang.ProfileInfo, error)
}

type UUIDsWorker struct {
	MojangUuidsProvider
}

func (ctx *UUIDsWorker) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.Handle("/mojang-uuid/{username}", http.HandlerFunc(ctx.getUUIDHandler)).Methods("GET")

	return router
}

func (ctx *UUIDsWorker) getUUIDHandler(response http.ResponseWriter, request *http.Request) {
	username := mux.Vars(request)["username"]
	profile, err := ctx.GetUuid(username)
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
