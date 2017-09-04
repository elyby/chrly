package http

import (
	"bytes"
	"image"
	"image/png"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"elyby/minecraft-skinsystem/db"
	"elyby/minecraft-skinsystem/model"
)

func TestConfig_Cape(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, _, capesRepo, wd := setupMocks(ctrl)

	cape := createCape()

	capesRepo.EXPECT().FindByUsername("mocked_username").Return(&model.Cape{
		File: bytes.NewReader(cape),
	}, nil)
	wd.EXPECT().IncCounter("capes.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/cloaks/mocked_username", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(200, resp.StatusCode)
	responseData, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(cape, responseData)
	assert.Equal("image/png", resp.Header.Get("Content-Type"))
}

func TestConfig_Cape2(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, _, capesRepo, wd := setupMocks(ctrl)

	capesRepo.EXPECT().FindByUsername("notch").Return(nil, &db.CapeNotFoundError{"notch"})
	wd.EXPECT().IncCounter("capes.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/cloaks/notch", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://skins.minecraft.net/MinecraftCloaks/notch.png", resp.Header.Get("Location"))
}

func TestConfig_CapeGET(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, _, capesRepo, wd := setupMocks(ctrl)

	cape := createCape()

	capesRepo.EXPECT().FindByUsername("mocked_username").Return(&model.Cape{
		File: bytes.NewReader(cape),
	}, nil)
	wd.EXPECT().IncCounter("capes.request", int64(1)).Times(0)
	wd.EXPECT().IncCounter("capes.get_request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/cloaks?name=mocked_username", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(200, resp.StatusCode)
	responseData, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(cape, responseData)
	assert.Equal("image/png", resp.Header.Get("Content-Type"))
}

func TestConfig_CapeGET2(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, _, capesRepo, wd := setupMocks(ctrl)

	capesRepo.EXPECT().FindByUsername("notch").Return(nil, &db.CapeNotFoundError{"notch"})
	wd.EXPECT().IncCounter("capes.request", int64(1)).Times(0)
	wd.EXPECT().IncCounter("capes.get_request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/cloaks?name=notch", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://skins.minecraft.net/MinecraftCloaks/notch.png", resp.Header.Get("Location"))
}

func TestConfig_CapeGET3(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/cloaks/?name=notch", nil)
	w := httptest.NewRecorder()

	(&Config{}).CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://skinsystem.ely.by/cloaks?name=notch", resp.Header.Get("Location"))
}

// Cape md5: 424ff79dce9940af89c28ad80de8aaad
func createCape() []byte {
	img := image.NewAlpha(image.Rect(0, 0, 64, 32))
	writer := &bytes.Buffer{}
	png.Encode(writer, img)

	pngBytes, _ := ioutil.ReadAll(writer)

	return pngBytes
}
