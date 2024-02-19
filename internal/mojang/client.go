package mojang

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type MojangApi struct {
	http          *http.Client
	batchUuidsUrl string
	profileUrl    string
}

func NewMojangApi(
	http *http.Client,
	batchUuidsUrl string,
	profileUrl string,
) *MojangApi {
	if batchUuidsUrl == "" {
		batchUuidsUrl = "https://api.mojang.com/profiles/minecraft"
	}

	if profileUrl == "" {
		profileUrl = "https://sessionserver.mojang.com/session/minecraft/profile/"
	}

	if !strings.HasSuffix(profileUrl, "/") {
		profileUrl += "/"
	}

	return &MojangApi{
		http,
		batchUuidsUrl,
		profileUrl,
	}
}

// Exchanges usernames array to array of uuids
// See https://wiki.vg/Mojang_API#Playernames_-.3E_UUIDs
func (c *MojangApi) UsernamesToUuids(ctx context.Context, usernames []string) ([]*ProfileInfo, error) {
	requestBody, _ := json.Marshal(usernames)
	request, err := http.NewRequestWithContext(ctx, "POST", c.batchUuidsUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errorFromResponse(response)
	}

	var result []*ProfileInfo

	body, _ := io.ReadAll(response.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Obtains textures information for provided uuid
// See https://wiki.vg/Mojang_API#UUID_-.3E_Profile_.2B_Skin.2FCape
func (c *MojangApi) UuidToTextures(ctx context.Context, uuid string, signed bool) (*ProfileResponse, error) {
	// TODO: normalize request url for tracing
	normalizedUuid := strings.ReplaceAll(uuid, "-", "")
	url := c.profileUrl + normalizedUuid
	if signed {
		url += "?unsigned=false"
	}

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == 204 {
		return nil, nil
	}

	if response.StatusCode != 200 {
		return nil, errorFromResponse(response)
	}

	var result *ProfileResponse

	body, _ := io.ReadAll(response.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type ProfileResponse struct {
	Id    string      `json:"id"`
	Name  string      `json:"name"`
	Props []*Property `json:"properties"`

	once            sync.Once
	decodedTextures *TexturesProp
	decodedErr      error
}

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

func (t *ProfileResponse) DecodeTextures() (*TexturesProp, error) {
	t.once.Do(func() {
		var texturesProp string
		for _, prop := range t.Props {
			if prop.Name == "textures" {
				texturesProp = prop.Value
				break
			}
		}

		if texturesProp == "" {
			return
		}

		decodedTextures, err := DecodeTextures(texturesProp)
		if err != nil {
			t.decodedErr = err
		} else {
			t.decodedTextures = decodedTextures
		}
	})

	return t.decodedTextures, t.decodedErr
}

type Property struct {
	Name      string `json:"name"`
	Signature string `json:"signature,omitempty"`
	Value     string `json:"value"`
}

type ProfileInfo struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	IsLegacy bool   `json:"legacy,omitempty"`
	IsDemo   bool   `json:"demo,omitempty"`
}

func errorFromResponse(response *http.Response) error {
	switch {
	case response.StatusCode == 400:
		type errorResponse struct {
			Error   string `json:"error"`
			Message string `json:"errorMessage"`
		}

		var decodedError *errorResponse
		body, _ := io.ReadAll(response.Body)
		_ = json.Unmarshal(body, &decodedError)

		return &BadRequestError{ErrorType: decodedError.Error, Message: decodedError.Message}
	case response.StatusCode == 403:
		return &ForbiddenError{}
	case response.StatusCode == 429:
		return &TooManyRequestsError{}
	case response.StatusCode >= 500:
		return &ServerError{Status: response.StatusCode}
	}

	return fmt.Errorf("unexpected response status code: %d", response.StatusCode)
}

// When passed request params are invalid, Mojang returns 400 Bad Request error
type BadRequestError struct {
	ErrorType string
	Message   string
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("400 %s: %s", e.ErrorType, e.Message)
}

// When Mojang decides you're such a bad guy, this error appears (even if the request has no authorization)
type ForbiddenError struct {
}

func (*ForbiddenError) Error() string {
	return "403: Forbidden"
}

// When you exceed the set limit of requests, this error will be returned
type TooManyRequestsError struct {
}

func (*TooManyRequestsError) Error() string {
	return "429: Too Many Requests"
}

// ServerError happens when Mojang's API returns any response with 50* status
type ServerError struct {
	Status int
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("%d: %s", e.Status, "Server error")
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
