package ui

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Config struct {

}

func Start(cfg Config, s *uiService, lst net.Listener) {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/skins/{username}", s.Skin).Methods("GET")
	router.HandleFunc("/cloaks/{username}", s.Cape).Methods("GET")
	router.HandleFunc("/textures/{username}", s.Textures).Methods("GET")
	router.HandleFunc("/textures/signed/{username}", s.SignedTextures).Methods("GET")
	router.HandleFunc("/skins/{username}/face", s.Face).Methods("GET")
	router.HandleFunc("/skins/{username}/face.png", s.Face).Methods("GET")
	// Legacy
	router.HandleFunc("/minecraft.php", s.MinecraftPHP).Methods("GET")
	router.HandleFunc("/skins/", s.SkinGET).Methods("GET")
	router.HandleFunc("/cloaks/", s.CapeGET).Methods("GET")
	// 404
	router.NotFoundHandler = http.HandlerFunc(NotFound)

	server := &http.Server{
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 16,
		Handler: router,
	}

	go server.Serve(lst)
}
