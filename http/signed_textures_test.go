package http

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/db"
)

func TestConfig_SignedTextures(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, mocks := setupMocks(ctrl)

	mocks.Skins.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
	mocks.Log.EXPECT().IncCounter("signed_textures.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/textures/signed/mock_user", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(200, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))
	response, _ := ioutil.ReadAll(resp.Body)
	assert.JSONEq(`{
		"id": "0f657aa8bfbe415db7005750090d3af3",
		"name": "mock_user",
		"properties": [
			{
				"name": "textures",
				"signature": "mocked signature",
				"value": "mocked textures base64"
			},
			{
				"name": "ely",
				"value": "but why are you asking?"
			}
		]
	}`, string(response))
}

func TestConfig_SignedTextures2(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, mocks := setupMocks(ctrl)

	mocks.Skins.EXPECT().FindByUsername("mock_user").Return(nil, &db.SkinNotFoundError{})
	mocks.Log.EXPECT().IncCounter("signed_textures.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/textures/signed/mock_user", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(204, resp.StatusCode)
	response, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("", string(response))
}
