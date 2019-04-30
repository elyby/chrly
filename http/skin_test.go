package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/db"
	"github.com/elyby/chrly/model"
)

type skinsTestCase struct {
	Name                 string
	RequestUrl           string
	ExpectedLogKey       string
	ExistsInLocalStorage bool
	ExistsInMojang       bool
	HasSkinInMojangResp  bool
	AssertResponse       func(assert *testify.Assertions, resp *http.Response)
}

var skinsTestCases = []*skinsTestCase{
	{
		Name: "Obtain skin for known username",
		ExistsInLocalStorage: true,
		AssertResponse: func(assert *testify.Assertions, resp *http.Response) {
			assert.Equal(301, resp.StatusCode)
			assert.Equal("http://chrly/skin.png", resp.Header.Get("Location"))
		},
	},
	{
		Name: "Obtain skin for unknown username that exists in Mojang and has a cape",
		ExistsInLocalStorage: false,
		ExistsInMojang: true,
		HasSkinInMojangResp: true,
		AssertResponse: func(assert *testify.Assertions, resp *http.Response) {
			assert.Equal(301, resp.StatusCode)
			assert.Equal("http://mojang/skin.png", resp.Header.Get("Location"))
		},
	},
	{
		Name: "Obtain skin for unknown username that exists in Mojang, but don't has a cape",
		ExistsInLocalStorage: false,
		ExistsInMojang: true,
		HasSkinInMojangResp: false,
		AssertResponse: func(assert *testify.Assertions, resp *http.Response) {
			assert.Equal(404, resp.StatusCode)
		},
	},
	{
		Name: "Obtain skin for unknown username that doesn't exists in Mojang",
		ExistsInLocalStorage: false,
		ExistsInMojang: false,
		AssertResponse: func(assert *testify.Assertions, resp *http.Response) {
			assert.Equal(404, resp.StatusCode)
		},
	},
}

func TestConfig_Skin(t *testing.T) {
	performTest := func(t *testing.T, testCase *skinsTestCase) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		mocks.Log.EXPECT().IncCounter(testCase.ExpectedLogKey, int64(1))
		if testCase.ExistsInLocalStorage {
			mocks.Skins.EXPECT().FindByUsername("mock_username").Return(createSkinModel("mock_username", false), nil)
		} else {
			mocks.Skins.EXPECT().FindByUsername("mock_username").Return(nil, &db.SkinNotFoundError{Who: "mock_username"})
		}

		if testCase.ExistsInMojang {
			textures := createTexturesResponse(testCase.HasSkinInMojangResp, true)
			mocks.Queue.On("GetTexturesForUsername", "mock_username").Return(textures)
		} else {
			mocks.Queue.On("GetTexturesForUsername", "mock_username").Return(nil)
		}

		req := httptest.NewRequest("GET", testCase.RequestUrl, nil)
		w := httptest.NewRecorder()

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		testCase.AssertResponse(assert, resp)
	}

	t.Run("Normal API", func(t *testing.T) {
		for _, testCase := range skinsTestCases {
			testCase.RequestUrl = "http://chrly/skins/mock_username"
			testCase.ExpectedLogKey = "skins.request"
			t.Run(testCase.Name, func(t *testing.T) {
				performTest(t, testCase)
			})
		}
	})

	t.Run("GET fallback API", func(t *testing.T) {
		for _, testCase := range skinsTestCases {
			testCase.RequestUrl = "http://chrly/skins?name=mock_username"
			testCase.ExpectedLogKey = "skins.get_request"
			t.Run(testCase.Name, func(t *testing.T) {
				performTest(t, testCase)
			})
		}

		t.Run("Should trim trailing slash", func(t *testing.T) {
			assert := testify.New(t)

			req := httptest.NewRequest("GET", "http://chrly/skins/?name=notch", nil)
			w := httptest.NewRecorder()

			(&Config{}).CreateHandler().ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(301, resp.StatusCode)
			assert.Equal("http://chrly/skins?name=notch", resp.Header.Get("Location"))
		})

		t.Run("Return error when name is not provided", func(t *testing.T) {
			assert := testify.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			config, mocks := setupMocks(ctrl)
			mocks.Log.EXPECT().IncCounter("skins.get_request", int64(1))

			req := httptest.NewRequest("GET", "http://chrly/skins", nil)
			w := httptest.NewRecorder()

			config.CreateHandler().ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(400, resp.StatusCode)
		})
	})
}

func createSkinModel(username string, isSlim bool) *model.Skin {
	return &model.Skin{
		UserId:          1,
		Username:        username,
		Uuid:            "0f657aa8-bfbe-415d-b700-5750090d3af3", // Use non nil UUID to pass validation in api tests
		SkinId:          1,
		Hash:            "00000000000000000000000000000000",
		Url:             "http://chrly/skin.png",
		MojangTextures:  "mocked textures base64",
		MojangSignature: "mocked signature",
		IsSlim:          isSlim,
	}
}
