package mojang

import (
	"encoding/base64"
	"encoding/json"
)

type TexturesProp struct {
	Timestamp   int64             `json:"timestamp"`
	ProfileID   string            `json:"profileId"`
	ProfileName string            `json:"profileName"`
	Textures    *TexturesResponse `json:"textures"`
}

type TexturesResponse struct {
	Skin *SkinTexturesResponse `json:"SKIN,omitempty"`
	Cape *CapeTexturesResponse `json:"CAPE,omitempty"`
}

type SkinTexturesResponse struct {
	Url      string                `json:"url"`
	Metadata *SkinTexturesMetadata `json:"metadata,omitempty"`
}

type SkinTexturesMetadata struct {
	Model string `json:"model"`
}

type CapeTexturesResponse struct {
	Url string `json:"url"`
}

func DecodeTextures(encodedTextures string) (*TexturesProp, error) {
	jsonStr, err := base64.URLEncoding.DecodeString(encodedTextures)
	if err != nil {
		return nil, err
	}

	var result *TexturesProp
	err = json.Unmarshal(jsonStr, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func EncodeTextures(textures *TexturesProp) string {
	jsonSerialized, _ := json.Marshal(textures)
	return base64.URLEncoding.EncodeToString(jsonSerialized)
}
