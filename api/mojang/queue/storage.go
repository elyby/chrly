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

// SplittedStorage allows you to use separate storage engines to satisfy
// the Storage interface
type SplittedStorage struct {
	UuidsStorage
	TexturesStorage
}

func (s *SplittedStorage) GetUuid(username string) (string, error) {
	return s.UuidsStorage.GetUuid(username)
}

func (s *SplittedStorage) StoreUuid(username string, uuid string) {
	s.UuidsStorage.StoreUuid(username, uuid)
}

func (s *SplittedStorage) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	return s.TexturesStorage.GetTextures(uuid)
}

func (s *SplittedStorage) StoreTextures(textures *mojang.SignedTexturesResponse) {
	s.TexturesStorage.StoreTextures(textures)
}

// This error can be used to indicate, that requested
// value doesn't exists in the storage
type ValueNotFound struct {
}

func (*ValueNotFound) Error() string {
	return "value not found in the storage"
}
