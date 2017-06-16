package worker

import (
	"strconv"
	"elyby/minecraft-skinsystem/lib/external/accounts"
)

var AccountsTokenConfig *accounts.TokenRequest

var token *accounts.Token

const repeatsLimit = 3
var repeatsCount = 0

func getById(id int) (accounts.AccountInfoResponse, error) {
	return _getByField("id", strconv.Itoa(id))
}

func _getByField(field string, value string) (accounts.AccountInfoResponse, error) {
	defer resetRepeatsCount()

	apiToken, err := getToken()
	if err != nil {
		return accounts.AccountInfoResponse{}, err
	}

	result, err := apiToken.AccountInfo(field, value)
	if err != nil {
		_, ok := err.(*accounts.UnauthorizedResponse)
		if !ok || repeatsCount >= repeatsLimit {
			return accounts.AccountInfoResponse{}, err
		}

		repeatsCount++
		token = nil

		return _getByField(field, value)
	}

	return result, nil
}

func getToken() (*accounts.Token, error) {
	if token == nil {
		println("token is nil, trying to obtain new one")
		tempToken, err := accounts.GetToken(*AccountsTokenConfig)
		if err != nil {
			println("cannot obtain new one token", err)
			return &accounts.Token{}, err
		}

		token = &tempToken
	}

	return token, nil
}

func resetRepeatsCount() {
	repeatsCount = 0
}
