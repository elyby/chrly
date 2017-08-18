package accounts

import (
	"net/http"
	"strings"
	"testing"

	testify "github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

var config = &Config{
	Addr: "https://account.ely.by",
	Id: "mock-id",
	Secret: "mock-secret",
	Scopes: []string{"scope1", "scope2"},
}

func TestConfig_GetTokenWithAutoRefresh(t *testing.T) {
	assert := testify.New(t)

	testConfig := &Config{}
	*testConfig = *config

	result := testConfig.GetTokenWithAutoRefresh()
	assert.Equal(testConfig, result.config)
}

func TestAutoRefresh_AccountInfo(t *testing.T) {
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

	gock.New("https://account.ely.by").
		Get("/api/internal/accounts/info").
		Times(2).
		MatchParam("id", "1").
		MatchHeader("Authorization", "Bearer mocked-token").
		Reply(200).
		JSON(map[string]interface{}{
			"id": 1,
			"uuid": "0f657aa8-bfbe-415d-b700-5750090d3af3",
			"username": "dummy",
			"email": "dummy@ely.by",
		})

	client := &http.Client{}
	gock.InterceptClient(client)

	testConfig := &Config{}
	*testConfig = *config
	testConfig.Client = client

	autoRefresher := testConfig.GetTokenWithAutoRefresh()
	result, err := autoRefresher.AccountInfo("id", "1")
	if assert.NoError(err) {
		assert.Equal(1, result.Id)
		assert.Equal("0f657aa8-bfbe-415d-b700-5750090d3af3", result.Uuid)
		assert.Equal("dummy", result.Username)
		assert.Equal("dummy@ely.by", result.Email)
	}

	result2, err2 := autoRefresher.AccountInfo("id", "1")
	if assert.NoError(err2) {
		assert.Equal(result, result2, "Results should still be same without token refreshing")
	}
}

func TestAutoRefresh_AccountInfo2(t *testing.T) {
	assert := testify.New(t)

	defer gock.Off()
	gock.New("https://account.ely.by").
		Post("/api/oauth2/v1/token").
		Body(strings.NewReader("client_id=mock-id&client_secret=mock-secret&grant_type=client_credentials&scope=scope1%2Cscope2")).
		Reply(200).
		JSON(map[string]interface{}{
			"access_token": "mocked-token-1",
			"token_type": "Bearer",
			"expires_in": 86400,
		})

	gock.New("https://account.ely.by").
		Get("/api/internal/accounts/info").
		MatchParam("id", "1").
		MatchHeader("Authorization", "Bearer mocked-token-1").
		Reply(200).
		JSON(map[string]interface{}{
			"id": 1,
			"uuid": "0f657aa8-bfbe-415d-b700-5750090d3af3",
			"username": "dummy",
			"email": "dummy@ely.by",
		})

	gock.New("https://account.ely.by").
		Get("/api/internal/accounts/info").
		MatchParam("id", "1").
		MatchHeader("Authorization", "Bearer mocked-token-1").
		Reply(401).
		JSON(map[string]interface{}{
			"name": "Unauthorized",
			"message": "Incorrect token",
			"code": 0,
			"status": 401,
		})

	gock.New("https://account.ely.by").
		Post("/api/oauth2/v1/token").
		Body(strings.NewReader("client_id=mock-id&client_secret=mock-secret&grant_type=client_credentials&scope=scope1%2Cscope2")).
		Reply(200).
		JSON(map[string]interface{}{
			"access_token": "mocked-token-2",
			"token_type": "Bearer",
			"expires_in": 86400,
		})

	gock.New("https://account.ely.by").
		Get("/api/internal/accounts/info").
		MatchParam("id", "1").
		MatchHeader("Authorization", "Bearer mocked-token-2").
		Reply(200).
		JSON(map[string]interface{}{
			"id": 1,
			"uuid": "0f657aa8-bfbe-415d-b700-5750090d3af3",
			"username": "dummy",
			"email": "dummy@ely.by",
		})

	client := &http.Client{}
	gock.InterceptClient(client)

	testConfig := &Config{}
	*testConfig = *config
	testConfig.Client = client

	autoRefresher := testConfig.GetTokenWithAutoRefresh()
	result, err := autoRefresher.AccountInfo("id", "1")
	if assert.NoError(err) {
		assert.Equal(1, result.Id)
		assert.Equal("0f657aa8-bfbe-415d-b700-5750090d3af3", result.Uuid)
		assert.Equal("dummy", result.Username)
		assert.Equal("dummy@ely.by", result.Email)
	}

	result2, err2 := autoRefresher.AccountInfo("id", "1")
	if assert.NoError(err2) {
		assert.Equal(result, result2, "Results should still be same with refreshed token")
	}
}

func TestAutoRefresh_AccountInfo3(t *testing.T) {
	assert := testify.New(t)

	defer gock.Off()
	gock.New("https://account.ely.by").
		Post("/api/oauth2/v1/token").
		Body(strings.NewReader("client_id=mock-id&client_secret=mock-secret&grant_type=client_credentials&scope=scope1%2Cscope2")).
		Reply(200).
		JSON(map[string]interface{}{
			"access_token": "mocked-token-1",
			"token_type": "Bearer",
			"expires_in": 86400,
		})

	gock.New("https://account.ely.by").
		Get("/api/internal/accounts/info").
		MatchParam("id", "1").
		MatchHeader("Authorization", "Bearer mocked-token-1").
		Reply(404).
		JSON(map[string]interface{}{
			"name": "Not Found",
			"message": "Page not found.",
			"code": 0,
			"status": 404,
		})

	client := &http.Client{}
	gock.InterceptClient(client)

	testConfig := &Config{}
	*testConfig = *config
	testConfig.Client = client

	autoRefresher := testConfig.GetTokenWithAutoRefresh()
	result, err := autoRefresher.AccountInfo("id", "1")
	assert.Nil(result)
	assert.Error(err)
	assert.IsType(&NotFoundResponse{}, err)
}

func TestAutoRefresh_AccountInfo4(t *testing.T) {
	assert := testify.New(t)

	defer gock.Off()
	gock.New("https://account.ely.by").
		Post("/api/oauth2/v1/token").
		Times(3).
		Body(strings.NewReader("client_id=mock-id&client_secret=mock-secret&grant_type=client_credentials&scope=scope1%2Cscope2")).
		Reply(200).
		JSON(map[string]interface{}{
			"access_token": "mocked-token-1",
			"token_type": "Bearer",
			"expires_in": 86400,
		})

	gock.New("https://account.ely.by").
		Get("/api/internal/accounts/info").
		Times(3).
		MatchParam("id", "1").
		MatchHeader("Authorization", "Bearer mocked-token-1").
		Reply(401).
		JSON(map[string]interface{}{
			"name": "Unauthorized",
			"message": "Incorrect token",
			"code": 0,
			"status": 401,
		})

	client := &http.Client{}
	gock.InterceptClient(client)

	testConfig := &Config{}
	*testConfig = *config
	testConfig.Client = client

	autoRefresher := testConfig.GetTokenWithAutoRefresh()
	result, err := autoRefresher.AccountInfo("id", "1")
	assert.Nil(result)
	assert.Error(err)
	if !assert.IsType(&UnauthorizedResponse{}, err) {
		t.Fatal(err)
	}
}
