package http

import (
	"testing"

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

	return &Config{
		SkinsRepo: skinsRepo,
		CapesRepo: capesRepo,
		Auth:      authChecker,
		Logger:    wd,
	}, &mocks{
		Skins: skinsRepo,
		Capes: capesRepo,
		Auth:  authChecker,
		Log:   wd,
	}
}
