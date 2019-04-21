package mojang

import (
	testify "github.com/stretchr/testify/assert"
	"testing"
)

type texturesTestCase struct {
	Name    string
	Encoded string
	Decoded *TexturesProp
}

var texturesTestCases = []*texturesTestCase{
	{
		Name:    "property without textures",
		Encoded: "eyJ0aW1lc3RhbXAiOjE1NTU4NTYwMTA0OTQsInByb2ZpbGVJZCI6IjNlM2VlNmMzNWFmYTQ4YWJiNjFlOGNkOGM0MmZjMGQ5IiwicHJvZmlsZU5hbWUiOiJFcmlja1NrcmF1Y2giLCJ0ZXh0dXJlcyI6e319",
		Decoded: &TexturesProp{
			ProfileID:   "3e3ee6c35afa48abb61e8cd8c42fc0d9",
			ProfileName: "ErickSkrauch",
			Timestamp:   int64(1555856010494),
			Textures:    &TexturesResponse{},
		},
	},
	{
		Name:    "property with classic skin textures",
		Encoded: "eyJ0aW1lc3RhbXAiOjE1NTU4NTYzMDc0MTIsInByb2ZpbGVJZCI6IjNlM2VlNmMzNWFmYTQ4YWJiNjFlOGNkOGM0MmZjMGQ5IiwicHJvZmlsZU5hbWUiOiJFcmlja1NrcmF1Y2giLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvZmMxNzU3NjMzN2ExMDZkOWMyMmFjNzgyZTM2MmMxNmM0ZTBlNDliZTUzZmFhNDE4NTdiZmYzMzJiNzc5MjgxZSJ9fX0=",
		Decoded: &TexturesProp{
			ProfileID:   "3e3ee6c35afa48abb61e8cd8c42fc0d9",
			ProfileName: "ErickSkrauch",
			Timestamp:   int64(1555856307412),
			Textures: &TexturesResponse{
				Skin: &SkinTexturesResponse{
					Url: "http://textures.minecraft.net/texture/fc17576337a106d9c22ac782e362c16c4e0e49be53faa41857bff332b779281e",
				},
			},
		},
	},
	{
		Name:    "property with alex skin textures",
		Encoded: "eyJ0aW1lc3RhbXAiOjE1NTU4NTY0OTQ3OTEsInByb2ZpbGVJZCI6IjNlM2VlNmMzNWFmYTQ4YWJiNjFlOGNkOGM0MmZjMGQ5IiwicHJvZmlsZU5hbWUiOiJFcmlja1NrcmF1Y2giLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvNjlmNzUzNWY4YzNhMjE1ZDFkZTc3MmIyODdmMTc3M2IzNTg5OGVmNzUyZDI2YmRkZjRhMjVhZGFiNjVjMTg1OSIsIm1ldGFkYXRhIjp7Im1vZGVsIjoic2xpbSJ9fX19",
		Decoded: &TexturesProp{
			ProfileID:   "3e3ee6c35afa48abb61e8cd8c42fc0d9",
			ProfileName: "ErickSkrauch",
			Timestamp:   int64(1555856494791),
			Textures: &TexturesResponse{
				Skin: &SkinTexturesResponse{
					Url: "http://textures.minecraft.net/texture/69f7535f8c3a215d1de772b287f1773b35898ef752d26bddf4a25adab65c1859",
					Metadata: &SkinTexturesMetadata{
						Model: "slim",
					},
				},
			},
		},
	},
	{
		Name:    "property with skin and cape textures",
		Encoded: "eyJ0aW1lc3RhbXAiOjE1NTU4NTc2NzUzMzUsInByb2ZpbGVJZCI6ImQ5MGI2OGJjODE3MjQzMjlhMDQ3ZjExODZkY2Q0MzM2IiwicHJvZmlsZU5hbWUiOiJha3Jvbm1hbjEiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvM2U2ZGVmY2I3ZGU1YTBlMDVjNzUyNWM2Y2Q0NmU0YjliNDE2YjkyZTBjZjRiYWExZTBhOWUyMTJhODg3ZjNmNyJ9LCJDQVBFIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvNzBlZmZmYWY4NmZlNWJjMDg5NjA4ZDNjYjI5N2QzZTI3NmI5ZWI3YThmOWYyZmU2NjU5YzIzYTJkOGIxOGVkZiJ9fX0=",
		Decoded: &TexturesProp{
			ProfileID:   "d90b68bc81724329a047f1186dcd4336",
			ProfileName: "akronman1",
			Timestamp:   int64(1555857675335),
			Textures: &TexturesResponse{
				Skin: &SkinTexturesResponse{
					Url: "http://textures.minecraft.net/texture/3e6defcb7de5a0e05c7525c6cd46e4b9b416b92e0cf4baa1e0a9e212a887f3f7",
				},
				Cape: &CapeTexturesResponse{
					Url: "http://textures.minecraft.net/texture/70efffaf86fe5bc089608d3cb297d3e276b9eb7a8f9f2fe6659c23a2d8b18edf",
				},
			},
		},
	},
}

func TestDecodeTextures(t *testing.T) {
	for _, testCase := range texturesTestCases {
		t.Run("decode "+testCase.Name, func(t *testing.T) {
			assert := testify.New(t)

			result, err := DecodeTextures(testCase.Encoded)
			assert.Nil(err)
			assert.Equal(testCase.Decoded, result)
		})
	}

	t.Run("should return error if invalid base64 passed", func(t *testing.T) {
		assert := testify.New(t)

		result, err := DecodeTextures("invalid base64")
		assert.Error(err)
		assert.Nil(result)
	})

	t.Run("should return error if invalid json found inside base64", func(t *testing.T) {
		assert := testify.New(t)

		result, err := DecodeTextures("aW52YWxpZCBqc29u") // encoded "invalid json"
		assert.Error(err)
		assert.Nil(result)
	})
}

func TestEncodeTextures(t *testing.T) {
	for _, testCase := range texturesTestCases {
		t.Run("encode "+testCase.Name, func(t *testing.T) {
			assert := testify.New(t)

			result := EncodeTextures(testCase.Decoded)
			assert.Equal(testCase.Encoded, result)
		})
	}
}
