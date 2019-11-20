package http

import (
	"bytes"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"github.com/elyby/chrly/db"
	"github.com/elyby/chrly/model"
)

type capesTestCase struct {
	Name                 string
	RequestUrl           string
	ExpectedLogKey       string
	ExistsInLocalStorage bool
	ExistsInMojang       bool
	HasCapeInMojangResp  bool
	AssertResponse       func(assert *testify.Assertions, resp *http.Response)
}

var capesTestCases = []*capesTestCase{
	{
		Name:                 "Obtain cape for known username",
		ExistsInLocalStorage: true,
		AssertResponse: func(assert *testify.Assertions, resp *http.Response) {
			assert.Equal(200, resp.StatusCode)
			responseData, _ := ioutil.ReadAll(resp.Body)
			assert.Equal(createCape(), responseData)
			assert.Equal("image/png", resp.Header.Get("Content-Type"))
		},
	},
	{
		Name:                 "Obtain cape for unknown username that exists in Mojang and has a cape",
		ExistsInLocalStorage: false,
		ExistsInMojang:       true,
		HasCapeInMojangResp:  true,
		AssertResponse: func(assert *testify.Assertions, resp *http.Response) {
			assert.Equal(301, resp.StatusCode)
			assert.Equal("http://mojang/cape.png", resp.Header.Get("Location"))
		},
	},
	{
		Name:                 "Obtain cape for unknown username that exists in Mojang, but don't has a cape",
		ExistsInLocalStorage: false,
		ExistsInMojang:       true,
		HasCapeInMojangResp:  false,
		AssertResponse: func(assert *testify.Assertions, resp *http.Response) {
			assert.Equal(404, resp.StatusCode)
		},
	},
	{
		Name:                 "Obtain cape for unknown username that doesn't exists in Mojang",
		ExistsInLocalStorage: false,
		ExistsInMojang:       false,
		AssertResponse: func(assert *testify.Assertions, resp *http.Response) {
			assert.Equal(404, resp.StatusCode)
		},
	},
}

func TestConfig_Cape(t *testing.T) {
	performTest := func(t *testing.T, testCase *capesTestCase) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		mocks.Log.EXPECT().IncCounter(testCase.ExpectedLogKey, int64(1))
		if testCase.ExistsInLocalStorage {
			mocks.Capes.EXPECT().FindByUsername("mock_username").Return(&model.Cape{
				File: bytes.NewReader(createCape()),
			}, nil)
		} else {
			mocks.Capes.EXPECT().FindByUsername("mock_username").Return(nil, &db.CapeNotFoundError{Who: "mock_username"})
		}

		if testCase.ExistsInMojang {
			textures := createTexturesResponse(false, testCase.HasCapeInMojangResp)
			mocks.MojangProvider.On("GetForUsername", "mock_username").Return(textures, nil)
		} else {
			mocks.MojangProvider.On("GetForUsername", "mock_username").Return(nil, nil)
		}

		req := httptest.NewRequest("GET", testCase.RequestUrl, nil)
		w := httptest.NewRecorder()

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		testCase.AssertResponse(assert, resp)
	}

	t.Run("Normal API", func(t *testing.T) {
		for _, testCase := range capesTestCases {
			testCase.RequestUrl = "http://chrly/cloaks/mock_username"
			testCase.ExpectedLogKey = "capes.request"
			t.Run(testCase.Name, func(t *testing.T) {
				performTest(t, testCase)
			})
		}
	})

	t.Run("GET fallback API", func(t *testing.T) {
		for _, testCase := range capesTestCases {
			testCase.RequestUrl = "http://chrly/cloaks?name=mock_username"
			testCase.ExpectedLogKey = "capes.get_request"
			t.Run(testCase.Name, func(t *testing.T) {
				performTest(t, testCase)
			})
		}

		t.Run("Should trim trailing slash", func(t *testing.T) {
			assert := testify.New(t)

			req := httptest.NewRequest("GET", "http://chrly/cloaks/?name=notch", nil)
			w := httptest.NewRecorder()

			(&Config{}).CreateHandler().ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(301, resp.StatusCode)
			assert.Equal("http://chrly/cloaks?name=notch", resp.Header.Get("Location"))
		})

		t.Run("Return error when name is not provided", func(t *testing.T) {
			assert := testify.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			config, mocks := setupMocks(ctrl)
			mocks.Log.EXPECT().IncCounter("capes.get_request", int64(1))

			req := httptest.NewRequest("GET", "http://chrly/cloaks", nil)
			w := httptest.NewRecorder()

			config.CreateHandler().ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(400, resp.StatusCode)
		})
	})
}

// Cape md5: 424ff79dce9940af89c28ad80de8aaad
func createCape() []byte {
	img := image.NewAlpha(image.Rect(0, 0, 64, 32))
	writer := &bytes.Buffer{}
	_ = png.Encode(writer, img)
	pngBytes, _ := ioutil.ReadAll(writer)

	return pngBytes
}
