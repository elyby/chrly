package http

import (
	"testing"
	"time"

	"github.com/elyby/chrly/api/mojang"

	"github.com/elyby/chrly/tests"
	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/interfaces/mock_interfaces"
	"github.com/elyby/chrly/interfaces/mock_wd"
)

func TestParseUsername(t *testing.T) {
	assert := testify.New(t)
	assert.Equal("test", parseUsername("test.png"), "Function should trim .png at end")
	assert.Equal("test", parseUsername("test"), "Function should return string itself, if it not contains .png at end")
}

type mocks struct {
	Skins *mock_interfaces.MockSkinsRepository
	Capes *mock_interfaces.MockCapesRepository
	Queue *tests.MojangTexturesQueueMock
	Auth  *mock_interfaces.MockAuthChecker
	Log   *mock_wd.MockWatchdog
}

func setupMocks(ctrl *gomock.Controller) (
	*Config,
	*mocks,
) {
	skinsRepo := mock_interfaces.NewMockSkinsRepository(ctrl)
	capesRepo := mock_interfaces.NewMockCapesRepository(ctrl)
	authChecker := mock_interfaces.NewMockAuthChecker(ctrl)
	wd := mock_wd.NewMockWatchdog(ctrl)
	texturesQueue := &tests.MojangTexturesQueueMock{}

	return &Config{
		SkinsRepo:           skinsRepo,
		CapesRepo:           capesRepo,
		Auth:                authChecker,
		MojangTexturesQueue: texturesQueue,
		Logger:              wd,
	}, &mocks{
		Skins: skinsRepo,
		Capes: capesRepo,
		Auth:  authChecker,
		Queue: texturesQueue,
		Log:   wd,
	}
}

func createTexturesResponse(includeSkin bool, includeCape bool) *mojang.SignedTexturesResponse {
	timeZone, _ := time.LoadLocation("Europe/Minsk")
	textures := &mojang.TexturesProp{
		Timestamp: time.Date(2019, 4, 27, 23, 56, 12, 0, timeZone).Unix(),
		ProfileID: "00000000000000000000000000000000",
		ProfileName: "mock_user",
		Textures: &mojang.TexturesResponse{},
	}

	if includeSkin {
		textures.Textures.Skin = &mojang.SkinTexturesResponse{
			Url: "http://mojang/skin.png",
		}
	}

	if includeCape {
		textures.Textures.Cape = &mojang.CapeTexturesResponse{
			Url: "http://mojang/cape.png",
		}
	}

	response := &mojang.SignedTexturesResponse{
		Id: "00000000000000000000000000000000",
		Name: "mock_user",
		Props: []*mojang.Property{
			{
				Name: "textures",
				Value: mojang.EncodeTextures(textures),
			},
		},
	}

	return response
}
