package mojangtextures

import (
	"github.com/elyby/chrly/api/mojang"
)

// UuidsStorage is a key-value storage of Mojang usernames pairs to its UUIDs,
// used to reduce the load on the account information queue
type UuidsStorage interface {
	// Since only primitive types are used in this method, you should return a special error ValueNotFound
	// to return the information that no error has occurred and username does not have uuid
	GetUuid(username string) (string, error)
	// An empty uuid value can be passed if the corresponding account has not been found
	StoreUuid(username string, uuid string) error
}

// TexturesStorage is a Mojang's textures storage, used as a values cache to avoid 429 errors
type TexturesStorage interface {
	// Error should not have nil value only if the repository failed to determine if there are any textures
	// for this uuid or not at all. If there is information about the absence of textures, nil nil should be returned
	GetTextures(uuid string) (*mojang.SignedTexturesResponse, error)
	// The nil value can be passed when there are no textures for the corresponding uuid and we know about it
	StoreTextures(uuid string, textures *mojang.SignedTexturesResponse)
}

type Storage interface {
	UuidsStorage
	TexturesStorage
}

// SeparatedStorage allows you to use separate storage engines to satisfy
// the Storage interface
type SeparatedStorage struct {
	UuidsStorage
	TexturesStorage
}

func (s *SeparatedStorage) GetUuid(username string) (string, error) {
	return s.UuidsStorage.GetUuid(username)
}

func (s *SeparatedStorage) StoreUuid(username string, uuid string) error {
	return s.UuidsStorage.StoreUuid(username, uuid)
}

func (s *SeparatedStorage) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	return s.TexturesStorage.GetTextures(uuid)
}

func (s *SeparatedStorage) StoreTextures(uuid string, textures *mojang.SignedTexturesResponse) {
	s.TexturesStorage.StoreTextures(uuid, textures)
}

// This error can be used to indicate, that requested
// value doesn't exists in the storage
type ValueNotFound struct {
}

func (*ValueNotFound) Error() string {
	return "value not found in the storage"
}
