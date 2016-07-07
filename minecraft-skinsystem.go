package main

import (
	"log"
	"runtime"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mediocregopher/radix.v2/redis"

	"elyby/minecraft-skinsystem/lib/routes"
	"elyby/minecraft-skinsystem/lib/services"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	client, redisErr := redis.Dial("tcp", "redis:6379")
	if redisErr != nil {
		log.Fatal("Redis unavailable")
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", routes.NotFound)
	router.HandleFunc("/skins/{username}", routes.Skin).Methods("GET").Name("skins")
	router.HandleFunc("/cloaks/{username}", routes.Cape).Methods("GET").Name("cloaks")
	router.HandleFunc("/textures/{username}", routes.Textures).Methods("GET").Name("textures")
	// Legacy
	router.HandleFunc("/minecraft.php", routes.MinecraftPHP).Methods("GET")
	router.HandleFunc("/skins/", routes.SkinGET).Methods("GET")
	router.HandleFunc("/cloaks/", routes.CapeGET).Methods("GET")

	// TODO: убрать этого, т.к. он стар
	router.HandleFunc("/system/setSkin", routes.SetSkin).Methods("POST")

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/user/{username}/skin", routes.SetSkin).Methods("POST")

	services.Redis = client
	services.Router = router

	log.Fatal(http.ListenAndServe(":80", router))
}
