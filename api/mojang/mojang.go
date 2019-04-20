package mojang

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var HttpClient = &http.Client{}

type SignedTexturesResponse struct {
	Id    string     `json:"id"`
	Name  string     `json:"name"`
	Props []Property `json:"properties"`
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
	request, err := http.NewRequest("POST", "https://api.mojang.com/profiles/minecraft", bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}

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

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

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
	case response.StatusCode == 429:
		return &TooManyRequestsError{}
	case response.StatusCode >= 500:
		return &ServerError{response.StatusCode}
	}

	return nil
}

// Mojang API doesn't return a 404 Not Found error for non-existent data identifiers
// Instead, they return 204 with an empty body
type EmptyResponse struct {
}

func (*EmptyResponse) Error() string {
	return "Empty Response"
}

// When you exceed the set limit of requests, this error will be returned
type TooManyRequestsError struct {
}

func (*TooManyRequestsError) Error() string {
	return "Too Many Requests"
}

// ServerError happens when Mojang's API returns any response with 50* status
type ServerError struct {
	Status int
}

func (e *ServerError) Error() string {
	return "Server error"
}
