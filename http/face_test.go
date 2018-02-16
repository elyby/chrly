package http

import (
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/db"
)

func TestConfig_Face(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, mocks := setupMocks(ctrl)

	mocks.Skins.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
	mocks.Log.EXPECT().IncCounter("faces.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/skins/mock_user/face.png", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://ely.by/minecraft/skin_buffer/faces/55d2a8848764f5ff04012cdb093458bd.png", resp.Header.Get("Location"))
}

func TestConfig_Face2(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config, mocks := setupMocks(ctrl)

	mocks.Skins.EXPECT().FindByUsername("mock_user").Return(nil, &db.SkinNotFoundError{"mock_user"})
	mocks.Log.EXPECT().IncCounter("faces.request", int64(1))

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/skins/mock_user/face.png", nil)
	w := httptest.NewRecorder()

	config.CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(301, resp.StatusCode)
	assert.Equal("http://ely.by/minecraft/skin_buffer/faces/default.png", resp.Header.Get("Location"))
}
