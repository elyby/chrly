package model

import "os"

type Cape struct {
	File *os.File
}

type CapesRepository interface {
	FindByUsername(username string) (Cape, error)
}
