package services

import (
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/gorilla/mux"
)

var RedisPool *pool.Pool

var Router *mux.Router
