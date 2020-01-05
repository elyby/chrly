package http

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mono83/slf/wd"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/mojangtextures"
)

type UuidsProvider interface {
	GetUuid(username string) (*mojang.ProfileInfo, error)
}

type UUIDsWorker struct {
	ListenSpec string

	UUIDsProvider mojangtextures.UUIDsProvider
	Logger        wd.Watchdog
}

func (ctx *UUIDsWorker) Run() error {
	ctx.Logger.Info(fmt.Sprintf("Starting the worker, HTTP on: %s\n", ctx.ListenSpec))

	listener, err := net.Listen("tcp", ctx.ListenSpec)
	if err != nil {
		return err
	}

	server := &http.Server{
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second, // TODO: should I adjust this values?
		MaxHeaderBytes: 1 << 16,
		Handler:        ctx.CreateHandler(),
	}

	// noinspection GoUnhandledErrorResult
	go server.Serve(listener)

	s := waitForSignal()
	ctx.Logger.Info(fmt.Sprintf("Got signal: %v, exiting.", s))

	return nil
}

func (ctx *UUIDsWorker) CreateHandler() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = http.HandlerFunc(NotFound)

	router.Handle("/api/worker/mojang-uuid/{username}", http.HandlerFunc(ctx.GetUUID)).Methods("GET")

	return router
}

func (ctx *UUIDsWorker) GetUUID(response http.ResponseWriter, request *http.Request) {
	username := parseUsername(mux.Vars(request)["username"])
	profile, err := ctx.UUIDsProvider.GetUuid(username)
	if err != nil {
		if _, ok := err.(*mojang.TooManyRequestsError); ok {
			ctx.Logger.Warning("Got 429 Too Many Requests")
			response.WriteHeader(http.StatusTooManyRequests)
			return
		}

		ctx.Logger.Warning("Got non success response: :err", wd.ErrParam(err))
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
