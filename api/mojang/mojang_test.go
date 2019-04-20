package mojang

import (
	"net/http"
	"testing"

	testify "github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestUsernamesToUuids(t *testing.T) {
	t.Run("exchange usernames to uuids", func(t *testing.T) {
		assert := testify.New(t)

		defer gock.Off()
		gock.New("https://api.mojang.com").
			Post("/profiles/minecraft").
			JSON([]string{"Thinkofdeath", "maksimkurb"}).
			Reply(200).
			JSON([]map[string]interface{}{
				{
					"id":     "4566e69fc90748ee8d71d7ba5aa00d20",
					"name":   "Thinkofdeath",
					"legacy": false,
					"demo":   true,
				},
				{
					"id":   "0d252b7218b648bfb86c2ae476954d32",
					"name": "maksimkurb",
					// There is no legacy or demo fields
				},
			})

		client := &http.Client{}
		gock.InterceptClient(client)

		HttpClient = client

		result, err := UsernamesToUuids([]string{"Thinkofdeath", "maksimkurb"})
		if assert.NoError(err) {
			assert.Len(result, 2)
			assert.Equal("4566e69fc90748ee8d71d7ba5aa00d20", result[0].Id)
			assert.Equal("Thinkofdeath", result[0].Name)
			assert.False(result[0].IsLegacy)
			assert.True(result[0].IsDemo)

			assert.Equal("0d252b7218b648bfb86c2ae476954d32", result[1].Id)
			assert.Equal("maksimkurb", result[1].Name)
			assert.False(result[1].IsLegacy)
			assert.False(result[1].IsDemo)
		}
	})

	t.Run("handle too many requests response", func(t *testing.T) {
		assert := testify.New(t)

		defer gock.Off()
		gock.New("https://api.mojang.com").
			Post("/profiles/minecraft").
			Reply(429).
			JSON(map[string]interface{}{
				"error":        "TooManyRequestsException",
				"errorMessage": "The client has sent too many requests within a certain amount of time",
			})

		client := &http.Client{}
		gock.InterceptClient(client)

		HttpClient = client

		result, err := UsernamesToUuids([]string{"Thinkofdeath", "maksimkurb"})
		assert.Nil(result)
		assert.IsType(&TooManyRequestsError{}, err)
		assert.EqualError(err, "Too Many Requests")
	})

	t.Run("handle server error", func(t *testing.T) {
		assert := testify.New(t)

		defer gock.Off()
		gock.New("https://api.mojang.com").
			Post("/profiles/minecraft").
			Reply(500).
			BodyString("500 Internal Server Error")

		client := &http.Client{}
		gock.InterceptClient(client)

		HttpClient = client

		result, err := UsernamesToUuids([]string{"Thinkofdeath", "maksimkurb"})
		assert.Nil(result)
		assert.IsType(&ServerError{}, err)
		assert.EqualError(err, "Server error")
		assert.Equal(500, err.(*ServerError).Status)
	})
}

func TestUuidToTextures(t *testing.T) {
	t.Run("obtain not signed textures", func(t *testing.T) {
		assert := testify.New(t)

		defer gock.Off()
		gock.New("https://sessionserver.mojang.com").
			Get("/session/minecraft/profile/4566e69fc90748ee8d71d7ba5aa00d20").
			Reply(200).
			JSON(map[string]interface{}{
				"id":   "4566e69fc90748ee8d71d7ba5aa00d20",
				"name": "Thinkofdeath",
				"properties": []interface{}{
					map[string]interface{}{
						"name":  "textures",
						"value": "eyJ0aW1lc3RhbXAiOjE1NDMxMDczMDExODUsInByb2ZpbGVJZCI6IjQ1NjZlNjlmYzkwNzQ4ZWU4ZDcxZDdiYTVhYTAwZDIwIiwicHJvZmlsZU5hbWUiOiJUaGlua29mZGVhdGgiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvNzRkMWUwOGIwYmI3ZTlmNTkwYWYyNzc1ODEyNWJiZWQxNzc4YWM2Y2VmNzI5YWVkZmNiOTYxM2U5OTExYWU3NSJ9LCJDQVBFIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvYjBjYzA4ODQwNzAwNDQ3MzIyZDk1M2EwMmI5NjVmMWQ2NWExM2E2MDNiZjY0YjE3YzgwM2MyMTQ0NmZlMTYzNSJ9fX0=",
					},
				},
			})

		client := &http.Client{}
		gock.InterceptClient(client)

		HttpClient = client

		result, err := UuidToTextures("4566e69fc90748ee8d71d7ba5aa00d20", false)
		if assert.NoError(err) {
			assert.Equal("4566e69fc90748ee8d71d7ba5aa00d20", result.Id)
			assert.Equal("Thinkofdeath", result.Name)
			assert.Equal(1, len(result.Props))
			assert.Equal("textures", result.Props[0].Name)
			assert.Equal(476, len(result.Props[0].Value))
			assert.Equal("", result.Props[0].Signature)
		}
	})

	t.Run("obtain signed textures", func(t *testing.T) {
		assert := testify.New(t)

		defer gock.Off()
		gock.New("https://sessionserver.mojang.com").
			Get("/session/minecraft/profile/4566e69fc90748ee8d71d7ba5aa00d20").
			MatchParam("unsigned", "false").
			Reply(200).
			JSON(map[string]interface{}{
				"id":   "4566e69fc90748ee8d71d7ba5aa00d20",
				"name": "Thinkofdeath",
				"properties": []interface{}{
					map[string]interface{}{
						"name":      "textures",
						"signature": "signature string",
						"value":     "eyJ0aW1lc3RhbXAiOjE1NDMxMDczMDExODUsInByb2ZpbGVJZCI6IjQ1NjZlNjlmYzkwNzQ4ZWU4ZDcxZDdiYTVhYTAwZDIwIiwicHJvZmlsZU5hbWUiOiJUaGlua29mZGVhdGgiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvNzRkMWUwOGIwYmI3ZTlmNTkwYWYyNzc1ODEyNWJiZWQxNzc4YWM2Y2VmNzI5YWVkZmNiOTYxM2U5OTExYWU3NSJ9LCJDQVBFIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvYjBjYzA4ODQwNzAwNDQ3MzIyZDk1M2EwMmI5NjVmMWQ2NWExM2E2MDNiZjY0YjE3YzgwM2MyMTQ0NmZlMTYzNSJ9fX0=",
					},
				},
			})

		client := &http.Client{}
		gock.InterceptClient(client)

		HttpClient = client

		result, err := UuidToTextures("4566e69fc90748ee8d71d7ba5aa00d20", true)
		if assert.NoError(err) {
			assert.Equal("4566e69fc90748ee8d71d7ba5aa00d20", result.Id)
			assert.Equal("Thinkofdeath", result.Name)
			assert.Equal(1, len(result.Props))
			assert.Equal("textures", result.Props[0].Name)
			assert.Equal(476, len(result.Props[0].Value))
			assert.Equal("signature string", result.Props[0].Signature)
		}
	})

	t.Run("handle empty response", func(t *testing.T) {
		assert := testify.New(t)

		defer gock.Off()
		gock.New("https://sessionserver.mojang.com").
			Get("/session/minecraft/profile/4566e69fc90748ee8d71d7ba5aa00d20").
			Reply(204).
			BodyString("")

		client := &http.Client{}
		gock.InterceptClient(client)

		HttpClient = client

		result, err := UuidToTextures("4566e69fc90748ee8d71d7ba5aa00d20", false)
		assert.Nil(result)
		assert.IsType(&EmptyResponse{}, err)
		assert.EqualError(err, "Empty Response")
	})

	t.Run("handle too many requests response", func(t *testing.T) {
		assert := testify.New(t)

		defer gock.Off()
		gock.New("https://sessionserver.mojang.com").
			Get("/session/minecraft/profile/4566e69fc90748ee8d71d7ba5aa00d20").
			Reply(429).
			JSON(map[string]interface{}{
				"error":        "TooManyRequestsException",
				"errorMessage": "The client has sent too many requests within a certain amount of time",
			})

		client := &http.Client{}
		gock.InterceptClient(client)

		HttpClient = client

		result, err := UuidToTextures("4566e69fc90748ee8d71d7ba5aa00d20", false)
		assert.Nil(result)
		assert.IsType(&TooManyRequestsError{}, err)
		assert.EqualError(err, "Too Many Requests")
	})

	t.Run("handle server error", func(t *testing.T) {
		assert := testify.New(t)

		defer gock.Off()
		gock.New("https://sessionserver.mojang.com").
			Get("/session/minecraft/profile/4566e69fc90748ee8d71d7ba5aa00d20").
			Reply(500).
			BodyString("500 Internal Server Error")

		client := &http.Client{}
		gock.InterceptClient(client)

		HttpClient = client

		result, err := UuidToTextures("4566e69fc90748ee8d71d7ba5aa00d20", false)
		assert.Nil(result)
		assert.IsType(&ServerError{}, err)
		assert.EqualError(err, "Server error")
		assert.Equal(500, err.(*ServerError).Status)
	})
}
