package services

import (
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/gorilla/mux"
)

var Redis *redis.Client

var Router *mux.Router
