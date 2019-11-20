package mojangtextures

import (
	"time"

	"github.com/mono83/slf/wd"

	"github.com/elyby/chrly/api/mojang"
)

var uuidToTextures = mojang.UuidToTextures

type MojangApiTexturesProvider struct {
	Logger wd.Watchdog
}

func (ctx *MojangApiTexturesProvider) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	ctx.Logger.IncCounter("mojang_textures.textures.request", 1)

	start := time.Now()
	result, err := uuidToTextures(uuid, true)
	ctx.Logger.RecordTimer("mojang_textures.textures.request_time", time.Since(start))

	return result, err
}
