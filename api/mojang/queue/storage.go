package queue

import "github.com/elyby/chrly/api/mojang"

type Storage interface {
	Get(username string) *mojang.SignedTexturesResponse
	Set(textures *mojang.SignedTexturesResponse)
}

// NilStorage used for testing purposes
type NilStorage struct {
}

func (*NilStorage) Get(username string) *mojang.SignedTexturesResponse {
	return nil
}

func (*NilStorage) Set(textures *mojang.SignedTexturesResponse) {
}
