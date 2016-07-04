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
	router.HandleFunc("/skins/{username}", routes.GetSkin)
	router.HandleFunc("/textures/{username}", routes.GetTextures)
	router.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	log.Fatal(http.ListenAndServe(":80", router))
}
