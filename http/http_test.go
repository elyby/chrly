package http

import (
	"testing"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"elyby/minecraft-skinsystem/http/mock_wd"
	"elyby/minecraft-skinsystem/interfaces/mock_interfaces"
)

func TestParseUsername(t *testing.T) {
	assert := testify.New(t)
	assert.Equal("test", parseUsername("test.png"), "Function should trim .png at end")
	assert.Equal("test", parseUsername("test"), "Function should return string itself, if it not contains .png at end")
}

func TestBuildElyUrl(t *testing.T) {
	assert := testify.New(t)
	assert.Equal("http://ely.by/route", buildElyUrl("/route"), "Function should add prefix to the provided relative url.")
	assert.Equal("http://ely.by/test/route", buildElyUrl("http://ely.by/test/route"), "Function should do not add prefix to the provided prefixed url.")
}

func setupMocks(ctrl *gomock.Controller) (
	*Config,
	*mock_interfaces.MockSkinsRepository,
	*mock_interfaces.MockCapesRepository,
	*mock_wd.MockWatchdog,
) {
	skinsRepo := mock_interfaces.NewMockSkinsRepository(ctrl)
	capesRepo := mock_interfaces.NewMockCapesRepository(ctrl)
	wd := mock_wd.NewMockWatchdog(ctrl)

	return &Config{
		SkinsRepo: skinsRepo,
		CapesRepo: capesRepo,
		Logger: wd,
	}, skinsRepo, capesRepo, wd
}
