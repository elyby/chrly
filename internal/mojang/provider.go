package mojang

import (
	"errors"
	"regexp"
	"strings"

	"github.com/brunomvsouza/singleflight"
)

var InvalidUsername = errors.New("the username passed doesn't meet Mojang's requirements")

// https://help.minecraft.net/hc/en-us/articles/4408950195341#h_01GE5JX1Z0CZ833A7S54Y195KV
var allowedUsernamesRegex = regexp.MustCompile(`(?i)^[0-9a-z_]{3,16}$`)

type UuidsProvider interface {
	GetUuid(username string) (*ProfileInfo, error)
}

type TexturesProvider interface {
	GetTextures(uuid string) (*ProfileResponse, error)
}

type MojangTexturesProvider struct {
	UuidsProvider
	TexturesProvider

	group singleflight.Group[string, *ProfileResponse]
}

func (p *MojangTexturesProvider) GetForUsername(username string) (*ProfileResponse, error) {
	if !allowedUsernamesRegex.MatchString(username) {
		return nil, InvalidUsername
	}

	username = strings.ToLower(username)

	result, err, _ := p.group.Do(username, func() (*ProfileResponse, error) {
		profile, err := p.UuidsProvider.GetUuid(username)
		if err != nil {
			return nil, err
		}

		if profile == nil {
			return nil, nil
		}

		return p.TexturesProvider.GetTextures(profile.Id)
	})

	return result, err
}

type NilProvider struct {
}

func (*NilProvider) GetForUsername(username string) (*ProfileResponse, error) {
	return nil, nil
}
