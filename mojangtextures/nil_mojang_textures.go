package mojangtextures

import (
	"github.com/elyby/chrly/api/mojang"
)

type NilProvider struct {
}

func (p *NilProvider) GetForUsername(username string) (*mojang.SignedTexturesResponse, error) {
	return nil, nil
}
