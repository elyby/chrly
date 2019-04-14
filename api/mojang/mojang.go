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

	if response.StatusCode == 429 {
		return nil, &TooManyRequestsError{}
	}

	var result []*ProfileInfo

	body, _ := ioutil.ReadAll(response.Body)
	_ = json.Unmarshal(body, &result)

	return result, nil
}

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

	if response.StatusCode == 429 {
		return nil, &TooManyRequestsError{}
	}

	var result *SignedTexturesResponse

	body, _ := ioutil.ReadAll(response.Body)
	_ = json.Unmarshal(body, &result)

	return result, nil
}

type TooManyRequestsError struct {
}

func (e *TooManyRequestsError) Error() string {
	return "Too Many Requests"
}
