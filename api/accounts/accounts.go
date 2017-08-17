package accounts

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type Config struct {
	Addr   string
	Id     string
	Secret string
	Scopes []string

	Client *http.Client
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	config      *Config
}

func (config *Config) GetToken() (*Token, error) {
	form := url.Values{}
	form.Add("client_id", config.Id)
	form.Add("client_secret", config.Secret)
	form.Add("grant_type", "client_credentials")
	form.Add("scope", strings.Join(config.Scopes, ","))

	response, err := config.getHttpClient().Post(config.getTokenUrl(), "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var result *Token
	responseError := handleResponse(response)
	if responseError != nil {
		return nil, responseError
	}

	body, _ := ioutil.ReadAll(response.Body)
	unmarshalError := json.Unmarshal(body, &result)
	if unmarshalError != nil {
		return nil, err
	}

	result.config = config

	return result, nil
}

func (config *Config) getTokenUrl() string {
	return concatenateHostAndPath(config.Addr, "/api/oauth2/v1/token")
}

func (config *Config) getHttpClient() *http.Client {
	if config.Client == nil {
		config.Client = &http.Client{}
	}

	return config.Client
}

type AccountInfoResponse struct {
	Id       int    `json:"id"`
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (token *Token) AccountInfo(attribute string, value string) (*AccountInfoResponse, error) {
	request := token.newRequest("GET", token.accountInfoUrl(), nil)

	query := request.URL.Query()
	query.Add(attribute, value)
	request.URL.RawQuery = query.Encode()

	response, err := token.config.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var info *AccountInfoResponse

	responseError := handleResponse(response)
	if responseError != nil {
		return nil, responseError
	}

	body, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &info)

	return info, nil
}

func (token *Token) accountInfoUrl() string {
	return concatenateHostAndPath(token.config.Addr, "/api/internal/accounts/info")
}

func (token *Token) newRequest(method string, urlStr string, body io.Reader) *http.Request {
	request, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		panic(err)
	}

	request.Header.Add("Authorization", "Bearer " + token.AccessToken)

	return request
}

func concatenateHostAndPath(host string, pathToJoin string) string {
	u, _ := url.Parse(host)
	u.Path = path.Join(u.Path, pathToJoin)

	return u.String()
}

type UnauthorizedResponse struct {}

func (err UnauthorizedResponse) Error() string {
	return "Unauthorized response"
}

type ForbiddenResponse struct {}

func (err ForbiddenResponse) Error() string {
	return "Forbidden response"
}

type NotFoundResponse struct {}

func (err NotFoundResponse) Error() string {
	return "Not found"
}

type NotSuccessResponse struct {
	StatusCode int
}

func (err NotSuccessResponse) Error() string {
	return fmt.Sprintf("Response code is \"%d\"", err.StatusCode)
}

func handleResponse(response *http.Response) error {
	switch status := response.StatusCode; status {
	case 200:
		return nil
	case 401:
		return &UnauthorizedResponse{}
	case 403:
		return &ForbiddenResponse{}
	case 404:
		return &NotFoundResponse{}
	default:
		return &NotSuccessResponse{status}
	}
}
