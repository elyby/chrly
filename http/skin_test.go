package http

import (
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/db"
	"github.com/elyby/chrly/model"
)

func TestConfig_Skin(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, mocks := setupMocks(ctrl)

	mocks.Skins.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
	mocks.Log.EXPECT().IncCounter("skins.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/skins/mock_user", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://ely.by/minecraft/skins/skin.png", resp.Header.Get("Location"))
}

func TestConfig_Skin2(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, mocks := setupMocks(ctrl)

	mocks.Skins.EXPECT().FindByUsername("notch").Return(nil, &db.SkinNotFoundError{"notch"})
	mocks.Log.EXPECT().IncCounter("skins.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/skins/notch", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://skins.minecraft.net/MinecraftSkins/notch.png", resp.Header.Get("Location"))
}

func TestConfig_SkinGET(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, mocks := setupMocks(ctrl)

	mocks.Skins.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
	mocks.Log.EXPECT().IncCounter("skins.get_request", int64(1))
	mocks.Log.EXPECT().IncCounter("skins.request", int64(1)).Times(0)

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/skins?name=mock_user", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://ely.by/minecraft/skins/skin.png", resp.Header.Get("Location"))
}

func TestConfig_SkinGET2(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, mocks := setupMocks(ctrl)

	mocks.Skins.EXPECT().FindByUsername("notch").Return(nil, &db.SkinNotFoundError{"notch"})
	mocks.Log.EXPECT().IncCounter("skins.get_request", int64(1))
	mocks.Log.EXPECT().IncCounter("skins.request", int64(1)).Times(0)

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/skins?name=notch", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://skins.minecraft.net/MinecraftSkins/notch.png", resp.Header.Get("Location"))
}

func TestConfig_SkinGET3(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/skins/?name=notch", nil)
	w := httptest.NewRecorder()

	(&Config{}).CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://skinsystem.ely.by/skins?name=notch", resp.Header.Get("Location"))
}

func createSkinModel(username string, isSlim bool) *model.Skin {
	return &model.Skin{
		UserId:          1,
		Username:        username,
		Uuid:            "0f657aa8-bfbe-415d-b700-5750090d3af3",
		SkinId:          1,
		Hash:            "55d2a8848764f5ff04012cdb093458bd",
		Url:             "http://ely.by/minecraft/skins/skin.png",
		MojangTextures:  "mocked textures base64",
		MojangSignature: "mocked signature",
		IsSlim:          isSlim,
	}
}
