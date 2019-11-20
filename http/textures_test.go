package http

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/db"
	"github.com/elyby/chrly/model"
)

func TestConfig_Textures(t *testing.T) {
	t.Run("Obtain textures for exists user with only default skin", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		mocks.Log.EXPECT().IncCounter("textures.request", int64(1))

		mocks.Skins.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
		mocks.Capes.EXPECT().FindByUsername("mock_user").Return(nil, &db.CapeNotFoundError{Who: "mock_user"})

		req := httptest.NewRequest("GET", "http://chrly/textures/mock_user", nil)
		w := httptest.NewRecorder()

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(200, resp.StatusCode)
		assert.Equal("application/json", resp.Header.Get("Content-Type"))
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`{
			"SKIN": {
				"url": "http://chrly/skin.png"
			}
		}`, string(response))
	})

	t.Run("Obtain textures for exists user with only slim skin", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		mocks.Log.EXPECT().IncCounter("textures.request", int64(1))

		mocks.Skins.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", true), nil)
		mocks.Capes.EXPECT().FindByUsername("mock_user").Return(nil, &db.CapeNotFoundError{Who: "mock_user"})

		req := httptest.NewRequest("GET", "http://chrly/textures/mock_user", nil)
		w := httptest.NewRecorder()

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(200, resp.StatusCode)
		assert.Equal("application/json", resp.Header.Get("Content-Type"))
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`{
			"SKIN": {
				"url": "http://chrly/skin.png",
				"metadata": {
					"model": "slim"
				}
			}
		}`, string(response))
	})

	t.Run("Obtain textures for exists user with only cape", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		mocks.Log.EXPECT().IncCounter("textures.request", int64(1))

		mocks.Skins.EXPECT().FindByUsername("mock_user").Return(nil, &db.SkinNotFoundError{Who: "mock_user"})
		mocks.Capes.EXPECT().FindByUsername("mock_user").Return(&model.Cape{File: bytes.NewReader(createCape())}, nil)

		req := httptest.NewRequest("GET", "http://chrly/textures/mock_user", nil)
		w := httptest.NewRecorder()

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(200, resp.StatusCode)
		assert.Equal("application/json", resp.Header.Get("Content-Type"))
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`{
			"CAPE": {
				"url": "http://chrly/cloaks/mock_user"
			}
		}`, string(response))
	})

	t.Run("Obtain textures for exists user with skin and cape", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		mocks.Log.EXPECT().IncCounter("textures.request", int64(1))

		mocks.Skins.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
		mocks.Capes.EXPECT().FindByUsername("mock_user").Return(&model.Cape{File: bytes.NewReader(createCape())}, nil)

		req := httptest.NewRequest("GET", "http://chrly/textures/mock_user", nil)
		w := httptest.NewRecorder()

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(200, resp.StatusCode)
		assert.Equal("application/json", resp.Header.Get("Content-Type"))
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`{
			"SKIN": {
				"url": "http://chrly/skin.png"
			},
			"CAPE": {
				"url": "http://chrly/cloaks/mock_user"
			}
		}`, string(response))
	})

	t.Run("Obtain textures for not exists user that exists in Mojang", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		mocks.Log.EXPECT().IncCounter("textures.request", int64(1))

		mocks.Skins.EXPECT().FindByUsername("mock_username").Return(nil, &db.SkinNotFoundError{})
		mocks.Capes.EXPECT().FindByUsername("mock_username").Return(nil, &db.CapeNotFoundError{})
		mocks.MojangProvider.On("GetForUsername", "mock_username").Once().Return(createTexturesResponse(true, true), nil)

		req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
		w := httptest.NewRecorder()

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(200, resp.StatusCode)
		assert.Equal("application/json", resp.Header.Get("Content-Type"))
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`{
			"SKIN": {
				"url": "http://mojang/skin.png"
			},
			"CAPE": {
				"url": "http://mojang/cape.png"
			}
		}`, string(response))
	})

	t.Run("Obtain textures for not exists user that not exists in Mojang too", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		mocks.Log.EXPECT().IncCounter("textures.request", int64(1))

		mocks.Skins.EXPECT().FindByUsername("mock_username").Return(nil, &db.SkinNotFoundError{})
		mocks.Capes.EXPECT().FindByUsername("mock_username").Return(nil, &db.CapeNotFoundError{})
		mocks.MojangProvider.On("GetForUsername", "mock_username").Once().Return(nil, nil)

		req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
		w := httptest.NewRecorder()

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(204, resp.StatusCode)
	})
}
