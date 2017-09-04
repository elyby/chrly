package accounts

type AutoRefresh struct {
	token 		 *Token
	config 		 *Config
	repeatsCount int
}

const repeatsLimit = 3

func (config *Config) GetTokenWithAutoRefresh() *AutoRefresh {
	return &AutoRefresh{
		config: config,
	}
}

func (refresher *AutoRefresh) AccountInfo(attribute string, value string) (*AccountInfoResponse, error) {
	defer refresher.resetRepeatsCount()

	apiToken, err := refresher.getToken()
	if err != nil {
		return nil, err
	}

	result, err := apiToken.AccountInfo(attribute, value)
	if err != nil {
		_, isTokenExpire := err.(*UnauthorizedResponse)
		if !isTokenExpire || refresher.repeatsCount >= repeatsLimit - 1 {
			return nil, err
		}

		refresher.repeatsCount++
		refresher.token = nil

		return refresher.AccountInfo(attribute, value)
	}

	return result, nil
}

func (refresher *AutoRefresh) getToken() (*Token, error) {
	if refresher.token == nil {
		newToken, err := refresher.config.GetToken()
		if err != nil {
			return nil, err
		}

		refresher.token = newToken
	}

	return refresher.token, nil
}

func (refresher *AutoRefresh) resetRepeatsCount() {
	refresher.repeatsCount = 0
}
