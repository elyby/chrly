package profiles

import (
	"errors"

	"github.com/elyby/chrly/internal/db"
	"github.com/elyby/chrly/internal/mojang"
)

type ProfilesFinder interface {
	FindProfileByUsername(username string) (*db.Profile, error)
}

type MojangProfilesProvider interface {
	GetForUsername(username string) (*mojang.ProfileResponse, error)
}

type Provider struct {
	ProfilesFinder
	MojangProfilesProvider
}

func (p *Provider) FindProfileByUsername(username string, allowProxy bool) (*db.Profile, error) {
	profile, err := p.ProfilesFinder.FindProfileByUsername(username)
	if err != nil {
		return nil, err
	}

	if profile != nil && (profile.SkinUrl != "" || profile.CapeUrl != "") {
		return profile, nil
	}

	if allowProxy {
		mojangProfile, err := p.MojangProfilesProvider.GetForUsername(username)
		// If we at least know something about the user,
		// then we can ignore an error and return profile without textures
		if err != nil && profile != nil {
			return profile, nil
		}

		if err != nil || mojangProfile == nil {
			if errors.Is(err, mojang.InvalidUsername) {
				return nil, nil
			}

			return nil, err
		}

		decodedTextures, err := mojangProfile.DecodeTextures()
		if err != nil {
			return nil, err
		}

		profile = &db.Profile{
			Uuid:     mojangProfile.Id,
			Username: mojangProfile.Name,
		}

		// There might be no textures property
		if decodedTextures != nil {
			if decodedTextures.Textures.Skin != nil {
				profile.SkinUrl = decodedTextures.Textures.Skin.Url
				if decodedTextures.Textures.Skin.Metadata != nil {
					profile.SkinModel = decodedTextures.Textures.Skin.Metadata.Model
				}
			}

			if decodedTextures.Textures.Cape != nil {
				profile.CapeUrl = decodedTextures.Textures.Cape.Url
			}
		}

		var texturesProp *mojang.Property
		for _, prop := range mojangProfile.Props {
			if prop.Name == "textures" {
				texturesProp = prop
				break
			}
		}

		if texturesProp != nil {
			profile.MojangTextures = texturesProp.Value
			profile.MojangSignature = texturesProp.Signature
		}
	}

	return profile, nil
}
