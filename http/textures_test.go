package http

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"elyby/minecraft-skinsystem/db"
	"elyby/minecraft-skinsystem/model"
)

func TestConfig_Textures(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, skinsRepo, capesRepo, wd := setupMocks(ctrl)

	skinsRepo.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
	capesRepo.EXPECT().FindByUsername("mock_user").Return(nil, &db.CapeNotFoundError{"mock_user"})
	wd.EXPECT().IncCounter("textures.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/textures/mock_user", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(200, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))
	response, _ := ioutil.ReadAll(resp.Body)
	assert.JSONEq(`{
		"SKIN": {
			"url": "http://ely.by/minecraft/skins/skin.png",
			"hash": "55d2a8848764f5ff04012cdb093458bd"
		}
	}`, string(response))
}

func TestConfig_Textures2(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, skinsRepo, capesRepo, wd := setupMocks(ctrl)

	skinsRepo.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", true), nil)
	capesRepo.EXPECT().FindByUsername("mock_user").Return(nil, &db.CapeNotFoundError{"mock_user"})
	wd.EXPECT().IncCounter("textures.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/textures/mock_user", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(200, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))
	response, _ := ioutil.ReadAll(resp.Body)
	assert.JSONEq(`{
		"SKIN": {
			"url": "http://ely.by/minecraft/skins/skin.png",
			"hash": "55d2a8848764f5ff04012cdb093458bd",
			"metadata": {
				"model": "slim"
			}
		}
	}`, string(response))
}

func TestConfig_Textures3(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, skinsRepo, capesRepo, wd := setupMocks(ctrl)

	skinsRepo.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
	capesRepo.EXPECT().FindByUsername("mock_user").Return(&model.Cape{
		File: bytes.NewReader(createCape()),
	}, nil)
	wd.EXPECT().IncCounter("textures.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/textures/mock_user", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(200, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))
	response, _ := ioutil.ReadAll(resp.Body)
	assert.JSONEq(`{
		"SKIN": {
			"url": "http://ely.by/minecraft/skins/skin.png",
			"hash": "55d2a8848764f5ff04012cdb093458bd"
		},
		"CAPE": {
			"url": "http://skinsystem.ely.by/cloaks/mock_user",
			"hash": "424ff79dce9940af89c28ad80de8aaad"
		}
	}`, string(response))
}

func TestConfig_Textures4(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, skinsRepo, capesRepo, wd := setupMocks(ctrl)

	skinsRepo.EXPECT().FindByUsername("notch").Return(nil, &db.SkinNotFoundError{})
	capesRepo.EXPECT().FindByUsername("notch").Return(nil, &db.CapeNotFoundError{})
	wd.EXPECT().IncCounter("textures.request", int64(1))
	timeNow = func() time.Time {
		return time.Date(2017, time.August, 20, 0, 15, 54, 0, time.UTC)
	}

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/textures/notch", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(200, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))
	response, _ := ioutil.ReadAll(resp.Body)
	assert.JSONEq(`{
		"SKIN": {
			"url": "http://skins.minecraft.net/MinecraftSkins/notch.png",
			"hash": "5923cf3f7fa170a279e4d7a9483cfc52"
		}
	}`, string(response))
}

func TestBuildNonElyTexturesHash(t *testing.T) {
	assert := testify.New(t)
	timeNow = func() time.Time {
		return time.Date(2017, time.November, 30, 16, 15, 34, 0, time.UTC)
	}

	assert.Equal("686d788a5353cb636e8fdff727634d88", buildNonElyTexturesHash("username"), "Function should return fixed hash by username-time pair")
	assert.Equal("fb876f761683a10accdb17d403cef64c", buildNonElyTexturesHash("another-username"), "Function should return fixed hash by username-time pair")

	timeNow = func() time.Time {
		return time.Date(2017, time.November, 30, 16, 20, 12, 0, time.UTC)
	}

	assert.Equal("686d788a5353cb636e8fdff727634d88", buildNonElyTexturesHash("username"), "Function should do not change it's value if hour the same")
	assert.Equal("fb876f761683a10accdb17d403cef64c", buildNonElyTexturesHash("another-username"), "Function should return fixed hash by username-time pair")

	timeNow = func() time.Time {
		return time.Date(2017, time.November, 30, 17, 1, 3, 0, time.UTC)
	}

	assert.Equal("42277892fd24bc0ed86285b3bb8b8fad", buildNonElyTexturesHash("username"), "Function should change it's value if hour changed")
}
