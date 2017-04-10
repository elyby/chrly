package accounts

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
)

type AccountInfoResponse struct {
	Id       int    `json:"id"`
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

const internalAccountInfoUrl = domain + "/api/internal/accounts/info"

func (token *Token) AccountInfo(attribute string, value string) (AccountInfoResponse, error) {
	request, err := http.NewRequest("GET", internalAccountInfoUrl, nil)
	request.Header.Add("Authorization", "Bearer " + token.AccessToken)
	query := request.URL.Query()
	query.Add(attribute, value)
	request.URL.RawQuery = query.Encode()

	response, err := Client.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	var info AccountInfoResponse

	responseError := handleResponse(response)
	if responseError != nil {
		return info, responseError
	}

	body, _ := ioutil.ReadAll(response.Body)
	println("Raw account info response is " + string(body))
	json.Unmarshal(body, &info)

	return info, nil
}
