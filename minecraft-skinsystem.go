package main

import (
	"log"
	"runtime"
	"time"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mediocregopher/radix.v2/redis"

	"elyby/minecraft-skinsystem/lib/routes"
	"elyby/minecraft-skinsystem/lib/services"
)

const redisString string = "redis:6379"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	client, redisErr := redis.Dial("tcp", redisString)
	if redisErr != nil {
		log.Fatal("Redis unavailable")
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/skins/{username}", routes.Skin).Methods("GET").Name("skins")
	router.HandleFunc("/cloaks/{username}", routes.Cape).Methods("GET").Name("cloaks")
	router.HandleFunc("/textures/{username}", routes.Textures).Methods("GET").Name("textures")
	// Legacy
	router.HandleFunc("/minecraft.php", routes.MinecraftPHP).Methods("GET")
	router.HandleFunc("/skins/", routes.SkinGET).Methods("GET")
	router.HandleFunc("/cloaks/", routes.CapeGET).Methods("GET")
	// 404
	router.NotFoundHandler = http.HandlerFunc(routes.NotFound)

	// TODO: убрать этого, т.к. он стар
	router.HandleFunc("/system/setSkin", routes.SetSkin).Methods("POST")

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/user/{username}/skin", routes.SetSkin).Methods("POST")

	services.Redis = client
	services.Router = router

	go func() {
		for {
			time.Sleep(5 * time.Second)

			resp := services.Redis.Cmd("PING")
			if (resp.Err != nil) {
				log.Println("Redis not pinged. Try to reconnect")
				newClient, redisErr := redis.Dial("tcp", redisString)
				if (redisErr != nil) {
					log.Println("Cannot reconnect to redis")
				} else {
					services.Redis = newClient
					log.Println("Reconnected")
				}
			}
		}
	}()

	log.Println("Started");
	log.Fatal(http.ListenAndServe(":80", router))
}
