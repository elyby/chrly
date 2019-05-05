package queue

import "github.com/elyby/chrly/api/mojang"

type UuidsStorage interface {
	GetUuid(username string) (string, error)
	StoreUuid(username string, uuid string)
}

// nil value can be passed to the storage to indicate that there is no textures
// for uuid and we know about it. Return err only in case, when storage completely
// unable to load any information about textures
type TexturesStorage interface {
	GetTextures(uuid string) (*mojang.SignedTexturesResponse, error)
	StoreTextures(uuid string, textures *mojang.SignedTexturesResponse)
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

func (s *SplittedStorage) StoreTextures(uuid string, textures *mojang.SignedTexturesResponse) {
	s.TexturesStorage.StoreTextures(uuid, textures)
}

// This error can be used to indicate, that requested
// value doesn't exists in the storage
type ValueNotFound struct {
}

func (*ValueNotFound) Error() string {
	return "value not found in the storage"
}
