package mojangtextures

import (
	"github.com/elyby/chrly/api/mojang"
)

var uuidToTextures = mojang.UuidToTextures

type MojangApiTexturesProvider struct {
	Emitter
}

func (ctx *MojangApiTexturesProvider) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	ctx.Emit("mojang_textures:mojang_api_textures_provider:before_request", uuid)
	result, err := uuidToTextures(uuid, true)
	ctx.Emit("mojang_textures:mojang_api_textures_provider:after_request", uuid, result, err)

	return result, err
}
