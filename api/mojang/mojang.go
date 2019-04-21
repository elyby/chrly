package mojang

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

var HttpClient = &http.Client{
	Timeout: 3 * time.Second,
}

type SignedTexturesResponse struct {
	Id    string      `json:"id"`
	Name  string      `json:"name"`
	Props []*Property `json:"properties"`
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

// Exchanges usernames array to array of uuids
// See https://wiki.vg/Mojang_API#Playernames_-.3E_UUIDs
func UsernamesToUuids(usernames []string) ([]*ProfileInfo, error) {
	requestBody, _ := json.Marshal(usernames)
	request, _ := http.NewRequest("POST", "https://api.mojang.com/profiles/minecraft", bytes.NewBuffer(requestBody))

	request.Header.Set("Content-Type", "application/json")

	response, err := HttpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if responseErr := validateResponse(response); responseErr != nil {
		return nil, responseErr
	}

	var result []*ProfileInfo

	body, _ := ioutil.ReadAll(response.Body)
	_ = json.Unmarshal(body, &result)

	return result, nil
}

// Obtains textures information for provided uuid
// See https://wiki.vg/Mojang_API#UUID_-.3E_Profile_.2B_Skin.2FCape
func UuidToTextures(uuid string, signed bool) (*SignedTexturesResponse, error) {
	url := "https://sessionserver.mojang.com/session/minecraft/profile/" + uuid
	if signed {
		url += "?unsigned=false"
	}

	request, _ := http.NewRequest("GET", url, nil)

	response, err := HttpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if responseErr := validateResponse(response); responseErr != nil {
		return nil, responseErr
	}

	var result *SignedTexturesResponse

	body, _ := ioutil.ReadAll(response.Body)
	_ = json.Unmarshal(body, &result)

	return result, nil
}

func validateResponse(response *http.Response) error {
	switch {
	case response.StatusCode == 204:
		return &EmptyResponse{}
	case response.StatusCode == 400:
		type errorResponse struct {
			Error   string `json:"error"`
			Message string `json:"errorMessage"`
		}

		var decodedError *errorResponse
		body, _ := ioutil.ReadAll(response.Body)
		_ = json.Unmarshal(body, &decodedError)

		return &BadRequestError{ErrorType: decodedError.Error, Message: decodedError.Message}
	case response.StatusCode == 429:
		return &TooManyRequestsError{}
	case response.StatusCode >= 500:
		return &ServerError{Status: response.StatusCode}
	}

	return nil
}

type ResponseError interface {
	IsMojangError() bool
}

// Mojang API doesn't return a 404 Not Found error for non-existent data identifiers
// Instead, they return 204 with an empty body
type EmptyResponse struct {
}

func (*EmptyResponse) Error() string {
	return "Empty Response"
}

func (*EmptyResponse) IsMojangError() bool {
	return true
}

// When passed request params are invalid, Mojang returns 400 Bad Request error
type BadRequestError struct {
	ResponseError
	ErrorType string
	Message   string
}

func (e *BadRequestError) Error() string {
	return e.Message
}

func (*BadRequestError) IsMojangError() bool {
	return true
}

// When you exceed the set limit of requests, this error will be returned
type TooManyRequestsError struct {
	ResponseError
}

func (*TooManyRequestsError) Error() string {
	return "Too Many Requests"
}

func (*TooManyRequestsError) IsMojangError() bool {
	return true
}

// ServerError happens when Mojang's API returns any response with 50* status
type ServerError struct {
	ResponseError
	Status int
}

func (e *ServerError) Error() string {
	return "Server error"
}

func (*ServerError) IsMojangError() bool {
	return true
}
