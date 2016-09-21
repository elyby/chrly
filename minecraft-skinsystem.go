package main

import (
	"log"
	"runtime"
	"time"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"github.com/mediocregopher/radix.v2/pool"

	"elyby/minecraft-skinsystem/lib/routes"
	"elyby/minecraft-skinsystem/lib/services"
	"elyby/minecraft-skinsystem/lib/worker"
)

const redisString string = "redis:6379"
const redisPoolSize int = 10

const rabbitmqString string = "amqp://ely-skinsystem-app:ely-skinsystem-app-password@rabbitmq:5672/%2fely"

func main() {
	log.Println("Starting...")

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Println("Connecting to redis")
	redisPool, redisErr := pool.New("tcp", redisString, redisPoolSize)
	if (redisErr != nil) {
		log.Fatal("Redis unavailable")
	}
	log.Println("Connected to redis")

	log.Println("Connecting to rabbitmq")
	// TODO: rabbitmq становится доступен не сразу. Нужно дождаться, пока он станет доступен, периодически повторяя запросы
	rabbitConnection, rabbitmqErr := amqp.Dial(rabbitmqString)
	if (rabbitmqErr != nil) {
		log.Fatalf("%s", rabbitmqErr)
	}
	log.Println("Connected to rabbitmq. Trying to open a channel")
	rabbitChannel, rabbitmqErr := rabbitConnection.Channel()
	if (rabbitmqErr != nil) {
		log.Fatalf("%s", rabbitmqErr)
	}
	log.Println("Connected to rabbitmq channel")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/skins/{username}", routes.Skin).Methods("GET").Name("skins")
	router.HandleFunc("/cloaks/{username}", routes.Cape).Methods("GET").Name("cloaks")
	router.HandleFunc("/textures/{username}", routes.Textures).Methods("GET").Name("textures")
	router.HandleFunc("/skins/{username}/face", routes.Face).Methods("GET").Name("faces")
	router.HandleFunc("/skins/{username}/face.png", routes.Face).Methods("GET").Name("faces")
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

	go func() {
		period := 5
		for {
			time.Sleep(time.Duration(period) * time.Second)

			resp := services.RedisPool.Cmd("PING")
			if (resp.Err == nil) {
				// Если редис успешно пинганулся, значит всё хорошо
				continue
			}

			log.Println("Redis not pinged. Try to reconnect")
			newPool, redisErr := pool.New("tcp", redisString, redisPoolSize)
			if (redisErr != nil) {
				log.Printf("Cannot reconnect to redis, waiting %d seconds\n", period)
			} else {
				services.RedisPool = newPool
				log.Println("Reconnected")
			}
		}
	}()

	go worker.Listen()

	log.Println("Started");
	log.Fatal(http.ListenAndServe(":80", router))
}
