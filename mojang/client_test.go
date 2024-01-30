package mojang

import (
	"net/http"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/suite"

	testify "github.com/stretchr/testify/assert"
)

type MojangApiSuite struct {
	suite.Suite
	api *MojangApi
}

func (s *MojangApiSuite) SetupTest() {
	httpClient := &http.Client{}
	gock.InterceptClient(httpClient)
	s.api = NewMojangApi(httpClient, "", "")
}

func (s *MojangApiSuite) TearDownTest() {
	gock.Off()
}

func (s *MojangApiSuite) TestUsernamesToUuidsSuccessfully() {
	gock.New("https://api.mojang.com").
		Post("/profiles/minecraft").
		JSON([]string{"Thinkofdeath", "maksimkurb"}).
		Reply(200).
		JSON([]map[string]any{
			{
				"id":     "4566e69fc90748ee8d71d7ba5aa00d20",
				"name":   "Thinkofdeath",
				"legacy": false,
				"demo":   true,
			},
			{
				"id":   "0d252b7218b648bfb86c2ae476954d32",
				"name": "maksimkurb",
				// There are no legacy or demo fields
			},
		})

	result, err := s.api.UsernamesToUuids([]string{"Thinkofdeath", "maksimkurb"})
	if s.Assert().NoError(err) {
		s.Assert().Len(result, 2)
		s.Assert().Equal("4566e69fc90748ee8d71d7ba5aa00d20", result[0].Id)
		s.Assert().Equal("Thinkofdeath", result[0].Name)
		s.Assert().False(result[0].IsLegacy)
		s.Assert().True(result[0].IsDemo)

		s.Assert().Equal("0d252b7218b648bfb86c2ae476954d32", result[1].Id)
		s.Assert().Equal("maksimkurb", result[1].Name)
		s.Assert().False(result[1].IsLegacy)
		s.Assert().False(result[1].IsDemo)
	}
}

func (s *MojangApiSuite) TestUsernamesToUuidsBadRequest() {
	gock.New("https://api.mojang.com").
		Post("/profiles/minecraft").
		Reply(400).
		JSON(map[string]any{
			"error":        "IllegalArgumentException",
			"errorMessage": "profileName can not be null or empty.",
		})

	result, err := s.api.UsernamesToUuids([]string{""})
	s.Assert().Nil(result)
	s.Assert().IsType(&BadRequestError{}, err)
	s.Assert().EqualError(err, "400 IllegalArgumentException: profileName can not be null or empty.")
}

func (s *MojangApiSuite) TestUsernamesToUuidsForbidden() {
	gock.New("https://api.mojang.com").
		Post("/profiles/minecraft").
		Reply(403).
		BodyString("just because")

	result, err := s.api.UsernamesToUuids([]string{"Thinkofdeath", "maksimkurb"})
	s.Assert().Nil(result)
	s.Assert().IsType(&ForbiddenError{}, err)
	s.Assert().EqualError(err, "403: Forbidden")
}

func (s *MojangApiSuite) TestUsernamesToUuidsTooManyRequests() {
	gock.New("https://api.mojang.com").
		Post("/profiles/minecraft").
		Reply(429).
		JSON(map[string]any{
			"error":        "TooManyRequestsException",
			"errorMessage": "The client has sent too many requests within a certain amount of time",
		})

	result, err := s.api.UsernamesToUuids([]string{"Thinkofdeath", "maksimkurb"})
	s.Assert().Nil(result)
	s.Assert().IsType(&TooManyRequestsError{}, err)
	s.Assert().EqualError(err, "429: Too Many Requests")
}

func (s *MojangApiSuite) TestUsernamesToUuidsServerError() {
	gock.New("https://api.mojang.com").
		Post("/profiles/minecraft").
		Reply(500).
		BodyString("500 Internal Server Error")

	result, err := s.api.UsernamesToUuids([]string{"Thinkofdeath", "maksimkurb"})
	s.Assert().Nil(result)
	s.Assert().IsType(&ServerError{}, err)
	s.Assert().EqualError(err, "500: Server error")
	s.Assert().Equal(500, err.(*ServerError).Status)
}

func (s *MojangApiSuite) TestUuidToTexturesSuccessfulResponse() {
	gock.New("https://sessionserver.mojang.com").
		Get("/session/minecraft/profile/4566e69fc90748ee8d71d7ba5aa00d20").
		Reply(200).
		JSON(map[string]any{
			"id":   "4566e69fc90748ee8d71d7ba5aa00d20",
			"name": "Thinkofdeath",
			"properties": []any{
				map[string]any{
					"name":  "textures",
					"value": "eyJ0aW1lc3RhbXAiOjE1NDMxMDczMDExODUsInByb2ZpbGVJZCI6IjQ1NjZlNjlmYzkwNzQ4ZWU4ZDcxZDdiYTVhYTAwZDIwIiwicHJvZmlsZU5hbWUiOiJUaGlua29mZGVhdGgiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvNzRkMWUwOGIwYmI3ZTlmNTkwYWYyNzc1ODEyNWJiZWQxNzc4YWM2Y2VmNzI5YWVkZmNiOTYxM2U5OTExYWU3NSJ9LCJDQVBFIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvYjBjYzA4ODQwNzAwNDQ3MzIyZDk1M2EwMmI5NjVmMWQ2NWExM2E2MDNiZjY0YjE3YzgwM2MyMTQ0NmZlMTYzNSJ9fX0=",
				},
			},
		})

	result, err := s.api.UuidToTextures("4566e69fc90748ee8d71d7ba5aa00d20", false)
	s.Assert().NoError(err)
	s.Assert().Equal("4566e69fc90748ee8d71d7ba5aa00d20", result.Id)
	s.Assert().Equal("Thinkofdeath", result.Name)
	s.Assert().Equal(1, len(result.Props))
	s.Assert().Equal("textures", result.Props[0].Name)
	s.Assert().Equal(476, len(result.Props[0].Value))
	s.Assert().Equal("", result.Props[0].Signature)
}

func (s *MojangApiSuite) TestUuidToTexturesEmptyResponse() {
	gock.New("https://sessionserver.mojang.com").
		Get("/session/minecraft/profile/4566e69fc90748ee8d71d7ba5aa00d20").
		Reply(204).
		BodyString("")

	result, err := s.api.UuidToTextures("4566e69fc90748ee8d71d7ba5aa00d20", false)
	s.Assert().Nil(result)
	s.Assert().NoError(err)
}

func (s *MojangApiSuite) TestUuidToTexturesTooManyRequests() {
	gock.New("https://sessionserver.mojang.com").
		Get("/session/minecraft/profile/4566e69fc90748ee8d71d7ba5aa00d20").
		Reply(429).
		JSON(map[string]any{
			"error":        "TooManyRequestsException",
			"errorMessage": "The client has sent too many requests within a certain amount of time",
		})

	result, err := s.api.UuidToTextures("4566e69fc90748ee8d71d7ba5aa00d20", false)
	s.Assert().Nil(result)
	s.Assert().IsType(&TooManyRequestsError{}, err)
	s.Assert().EqualError(err, "429: Too Many Requests")
}

func (s *MojangApiSuite) TestUuidToTexturesServerError() {
	gock.New("https://sessionserver.mojang.com").
		Get("/session/minecraft/profile/4566e69fc90748ee8d71d7ba5aa00d20").
		Reply(500).
		BodyString("500 Internal Server Error")

	result, err := s.api.UuidToTextures("4566e69fc90748ee8d71d7ba5aa00d20", false)
	s.Assert().Nil(result)
	s.Assert().IsType(&ServerError{}, err)
	s.Assert().EqualError(err, "500: Server error")
	s.Assert().Equal(500, err.(*ServerError).Status)
}

func TestMojangApi(t *testing.T) {
	suite.Run(t, new(MojangApiSuite))
}

func TestSignedTexturesResponse(t *testing.T) {
	t.Run("DecodeTextures", func(t *testing.T) {
		obj := &ProfileResponse{
			Id:   "00000000000000000000000000000000",
			Name: "mock",
			Props: []*Property{
				{
					Name:  "textures",
					Value: "eyJ0aW1lc3RhbXAiOjE1NTU4NTYzMDc0MTIsInByb2ZpbGVJZCI6IjNlM2VlNmMzNWFmYTQ4YWJiNjFlOGNkOGM0MmZjMGQ5IiwicHJvZmlsZU5hbWUiOiJFcmlja1NrcmF1Y2giLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvZmMxNzU3NjMzN2ExMDZkOWMyMmFjNzgyZTM2MmMxNmM0ZTBlNDliZTUzZmFhNDE4NTdiZmYzMzJiNzc5MjgxZSJ9fX0=",
				},
			},
		}
		textures, err := obj.DecodeTextures()
		testify.Nil(t, err)
		testify.Equal(t, "3e3ee6c35afa48abb61e8cd8c42fc0d9", textures.ProfileID)
	})

	t.Run("DecodedTextures without textures prop", func(t *testing.T) {
		obj := &ProfileResponse{
			Id:    "00000000000000000000000000000000",
			Name:  "mock",
			Props: []*Property{},
		}
		textures, err := obj.DecodeTextures()
		testify.Nil(t, err)
		testify.Nil(t, textures)
	})
}

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
