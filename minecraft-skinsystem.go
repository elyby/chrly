package main

import (
	"log"
	"runtime"
	//"time"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mediocregopher/radix.v2/pool"

	"elyby/minecraft-skinsystem/lib/routes"
	"elyby/minecraft-skinsystem/lib/services"
	//"github.com/mediocregopher/radix.v2/redis"

	"github.com/streadway/amqp"
	"elyby/minecraft-skinsystem/lib/worker"
)

const redisString string = "redis:6379"
const rabbitmqString string = "amqp://ely-skinsystem-app:ely-skinsystem-app-password@rabbitmq:5672/%2fely"

func main() {
	log.Println("Starting...")

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Println("Connecting to redis")
	redisPool, redisErr := pool.New("tcp", redisString, 10)
	if redisErr != nil {
		log.Fatal("Redis unavailable")
	}
	log.Println("Connected to redis")

	log.Println("Connecting to rabbitmq")
	// TODO: rabbitmq становится доступен не сразу. Нужно дождаться, пока он станет доступен, периодически повторяя запросы
	rabbitConnection, rabbitmqErr := amqp.Dial(rabbitmqString)
	if rabbitmqErr != nil {
		log.Fatalf("%s", rabbitmqErr)
	}
	log.Println("Connected to rabbitmq. Trying to open a channel")
	rabbitChannel, rabbitmqErr := rabbitConnection.Channel()
	if rabbitmqErr != nil {
		log.Fatalf("%s", rabbitmqErr)
	}
	log.Println("Connected to rabbitmq channel")

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

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/user/{username}/skin", routes.SetSkin).Methods("POST")

	services.RedisPool = redisPool
	services.RabbitMQChannel = rabbitChannel

	/*go func() {
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
	}()*/

	go worker.Listen()

	log.Println("Started");
	log.Fatal(http.ListenAndServe(":80", router))
}
