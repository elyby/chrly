package accounts

import (
	"strings"
	"net/url"
	"io/ioutil"
	"encoding/json"
)

type TokenRequest struct {
	Id     string
	Secret string
	Scopes []string
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

const tokenUrl = domain + "/api/oauth2/v1/token"

func GetToken(request TokenRequest) (Token, error) {
	form := url.Values{}
	form.Add("client_id", request.Id)
	form.Add("client_secret", request.Secret)
	form.Add("grant_type", "client_credentials")
	form.Add("scope", strings.Join(request.Scopes, ","))

	response, err := Client.Post(tokenUrl, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	var result Token
	responseError := handleResponse(response)
	if responseError != nil {
		return result, responseError
	}

	body, _ := ioutil.ReadAll(response.Body)

	json.Unmarshal(body, &result)

	return result, nil
}
