package accounts

import (
	"net/http"
	"strings"
	"testing"

	testify "github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestConfig_GetToken(t *testing.T) {
	assert := testify.New(t)

	defer gock.Off()
	gock.New("https://account.ely.by").
		Post("/api/oauth2/v1/token").
		Body(strings.NewReader("client_id=mock-id&client_secret=mock-secret&grant_type=client_credentials&scope=scope1%2Cscope2")).
		Reply(200).
		JSON(map[string]interface{}{
			"access_token": "mocked-token",
			"token_type": "Bearer",
			"expires_in": 86400,
		})

	client := &http.Client{}
	gock.InterceptClient(client)

	config := &Config{
		Addr: "https://account.ely.by",
		Id: "mock-id",
		Secret: "mock-secret",
		Scopes: []string{"scope1", "scope2"},
		Client: client,
	}

	result, err := config.GetToken()
	if assert.NoError(err) {
		assert.Equal("mocked-token", result.AccessToken)
		assert.Equal("Bearer", result.TokenType)
		assert.Equal(86400, result.ExpiresIn)
	}
}

func TestToken_AccountInfo(t *testing.T) {
	assert := testify.New(t)

	defer gock.Off()
	// To test valid behavior
	gock.New("https://account.ely.by").
		Get("/api/internal/accounts/info").
		MatchParam("id", "1").
		MatchHeader("Authorization", "Bearer mock-token").
		Reply(200).
		JSON(map[string]interface{}{
			"id": 1,
			"uuid": "0f657aa8-bfbe-415d-b700-5750090d3af3",
			"username": "dummy",
			"email": "dummy@ely.by",
		})

	// To test behavior on invalid or expired token
	gock.New("https://account.ely.by").
		Get("/api/internal/accounts/info").
		MatchParam("id", "1").
		MatchHeader("Authorization", "Bearer mock-token").
		Reply(401).
		JSON(map[string]interface{}{
			"name": "Unauthorized",
			"message": "Incorrect token",
			"code": 0,
			"status": 401,
		})

	client := &http.Client{}
	gock.InterceptClient(client)

	token := &Token{
		AccessToken: "mock-token",
		config: &Config{
			Addr: "https://account.ely.by",
			Client: client,
		},
	}

	result, err := token.AccountInfo("id", "1")
	if assert.NoError(err) {
		assert.Equal(1, result.Id)
		assert.Equal("0f657aa8-bfbe-415d-b700-5750090d3af3", result.Uuid)
		assert.Equal("dummy", result.Username)
		assert.Equal("dummy@ely.by", result.Email)
	}

	result2, err2 := token.AccountInfo("id", "1")
	assert.Nil(result2)
	assert.Error(err2)
	assert.IsType(&UnauthorizedResponse{}, err2)
}
