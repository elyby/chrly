package http

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/mono83/slf/wd"

	"elyby/minecraft-skinsystem/interfaces"
)

type Config struct {
	ListenSpec string

	SkinsRepo interfaces.SkinsRepository
	CapesRepo interfaces.CapesRepository
	Logger    wd.Watchdog
	Auth      interfaces.AuthChecker
}

func (cfg *Config) Run() error {
	cfg.Logger.Info(fmt.Sprintf("Starting, HTTP on: %s\n", cfg.ListenSpec))

	listener, err := net.Listen("tcp", cfg.ListenSpec)
	if err != nil {
		return err
	}

	server := &http.Server{
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 16,
		Handler:        cfg.CreateHandler(),
	}

	go server.Serve(listener)

	s := waitForSignal()
	cfg.Logger.Info(fmt.Sprintf("Got signal: %v, exiting.", s))

	return nil
}

func (cfg *Config) CreateHandler() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/skins/{username}", cfg.Skin).Methods("GET")
	router.HandleFunc("/cloaks/{username}", cfg.Cape).Methods("GET").Name("cloaks")
	router.HandleFunc("/textures/{username}", cfg.Textures).Methods("GET")
	router.HandleFunc("/textures/signed/{username}", cfg.SignedTextures).Methods("GET")
	router.HandleFunc("/skins/{username}/face", cfg.Face).Methods("GET")
	router.HandleFunc("/skins/{username}/face.png", cfg.Face).Methods("GET")
	// Legacy
	router.HandleFunc("/skins", cfg.SkinGET).Methods("GET")
	router.HandleFunc("/cloaks", cfg.CapeGET).Methods("GET")
	// API
	router.Handle("/api/skins", cfg.Authenticate(http.HandlerFunc(cfg.PostSkin))).Methods("POST")
	// 404
	router.NotFoundHandler = http.HandlerFunc(cfg.NotFound)

	return router
}

func parseUsername(username string) string {
	const suffix = ".png"
	if strings.HasSuffix(username, suffix) {
		username = strings.TrimSuffix(username, suffix)
	}

	return username
}

func buildElyUrl(route string) string {
	prefix := "http://ely.by"
	if !strings.HasPrefix(route, prefix) {
		route = prefix + route
	}

	return route
}

func waitForSignal() os.Signal {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	return <-ch
}
