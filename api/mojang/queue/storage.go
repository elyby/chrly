package queue

import "github.com/elyby/chrly/api/mojang"

type UuidsStorage interface {
	GetUuid(username string) (string, error)
	StoreUuid(username string, uuid string)
}

type TexturesStorage interface {
	// nil can be returned to indicate that there is no textures for uuid
	// and we know about it. Return err only in case, when storage completely
	// don't know anything about uuid
	GetTextures(uuid string) (*mojang.SignedTexturesResponse, error)
	StoreTextures(textures *mojang.SignedTexturesResponse)
}

type Storage interface {
	UuidsStorage
	TexturesStorage
}

// This error can be used to indicate, that requested
// value doesn't exists in the storage
type ValueNotFound struct {
}

func (*ValueNotFound) Error() string {
	return "value not found in storage"
}
