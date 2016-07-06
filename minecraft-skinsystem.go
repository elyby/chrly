package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mediocregopher/radix.v2/redis"

	"elyby/minecraft-skinsystem/lib/routes"
	"elyby/minecraft-skinsystem/lib/services"
)

var client, redisErr = redis.Dial("tcp", "redis:6379")

func main() {
	if redisErr != nil {
		log.Fatal("Redis unavailable")
	}

	services.Redis = client

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", routes.NotFound)
	router.HandleFunc("/skins/{username}", routes.Skin).Methods("GET")
	router.HandleFunc("/textures/{username}", routes.Textures).Methods("GET")
	router.HandleFunc("/system/setSkin", routes.SetSkin).Methods("POST") // TODO: убрать этого, т.к. он стар

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/user/{username}/skin", routes.SetSkin).Methods("POST")

	log.Fatal(http.ListenAndServe(":80", router))
}
