package http

import (
	"bytes"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/model"
)

/***************
 * Setup mocks *
 ***************/

type skinsRepositoryMock struct {
	mock.Mock
}

func (m *skinsRepositoryMock) FindByUsername(username string) (*model.Skin, error) {
	args := m.Called(username)
	var result *model.Skin
	if casted, ok := args.Get(0).(*model.Skin); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *skinsRepositoryMock) FindByUserId(id int) (*model.Skin, error) {
	args := m.Called(id)
	var result *model.Skin
	if casted, ok := args.Get(0).(*model.Skin); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *skinsRepositoryMock) Save(skin *model.Skin) error {
	args := m.Called(skin)
	return args.Error(0)
}

func (m *skinsRepositoryMock) RemoveByUserId(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *skinsRepositoryMock) RemoveByUsername(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

type capesRepositoryMock struct {
	mock.Mock
}

func (m *capesRepositoryMock) FindByUsername(username string) (*model.Cape, error) {
	args := m.Called(username)
	var result *model.Cape
	if casted, ok := args.Get(0).(*model.Cape); ok {
		result = casted
	}

	return result, args.Error(1)
}

type mojangTexturesProviderMock struct {
	mock.Mock
}

func (m *mojangTexturesProviderMock) GetForUsername(username string) (*mojang.SignedTexturesResponse, error) {
	args := m.Called(username)
	var result *mojang.SignedTexturesResponse
	if casted, ok := args.Get(0).(*mojang.SignedTexturesResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

type skinsystemTestSuite struct {
	suite.Suite

	App *Skinsystem

	SkinsRepository        *skinsRepositoryMock
	CapesRepository        *capesRepositoryMock
	MojangTexturesProvider *mojangTexturesProviderMock
	Emitter                *emitterMock
}

/********************
 * Setup test suite *
 ********************/

func (suite *skinsystemTestSuite) SetupTest() {
	suite.SkinsRepository = &skinsRepositoryMock{}
	suite.CapesRepository = &capesRepositoryMock{}
	suite.MojangTexturesProvider = &mojangTexturesProviderMock{}
	suite.Emitter = &emitterMock{}

	suite.App = &Skinsystem{
		SkinsRepo:              suite.SkinsRepository,
		CapesRepo:              suite.CapesRepository,
		MojangTexturesProvider: suite.MojangTexturesProvider,
		Emitter:                suite.Emitter,
	}
}

func (suite *skinsystemTestSuite) TearDownTest() {
	suite.SkinsRepository.AssertExpectations(suite.T())
	suite.CapesRepository.AssertExpectations(suite.T())
	suite.MojangTexturesProvider.AssertExpectations(suite.T())
	suite.Emitter.AssertExpectations(suite.T())
}

func (suite *skinsystemTestSuite) RunSubTest(name string, subTest func()) {
	suite.SetupTest()
	suite.Run(name, subTest)
	suite.TearDownTest()
}

/*************
 * Run tests *
 *************/

func TestSkinsystem(t *testing.T) {
	suite.Run(t, new(skinsystemTestSuite))
}

type skinsystemTestCase struct {
	Name       string
	BeforeTest func(suite *skinsystemTestSuite)
	AfterTest  func(suite *skinsystemTestSuite, response *http.Response)
}

/************************
 * Get skin tests cases *
 ************************/

var skinsTestsCases = []*skinsystemTestCase{
	{
		Name: "Username exists in the local storage",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(301, response.StatusCode)
			suite.Equal("http://chrly/skin.png", response.Header.Get("Location"))
		},
	},
	{
		Name: "Username doesn't exists on the local storage, but exists on Mojang and has textures",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(nil, &SkinNotFoundError{Who: "mock_username"})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponse(true, false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(301, response.StatusCode)
			suite.Equal("http://mojang/skin.png", response.Header.Get("Location"))
		},
	},
	{
		Name: "Username doesn't exists on the local storage, but exists on Mojang and has no textures",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(nil, &SkinNotFoundError{Who: "mock_username"})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponse(false, false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
	{
		Name: "Username doesn't exists on the local storage and doesn't exists on Mojang",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(nil, &SkinNotFoundError{Who: "mock_username"})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
}

func (suite *skinsystemTestSuite) TestSkin() {
	for _, testCase := range skinsTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/skins/mock_username", nil)
			w := httptest.NewRecorder()

			suite.App.Handler().ServeHTTP(w, req)

			testCase.AfterTest(suite, w.Result())
		})
	}

	suite.RunSubTest("Pass username with png extension", func() {
		suite.SkinsRepository.On("FindByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)

		req := httptest.NewRequest("GET", "http://chrly/skins/mock_username.png", nil)
		w := httptest.NewRecorder()

		suite.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		suite.Equal(301, resp.StatusCode)
		suite.Equal("http://chrly/skin.png", resp.Header.Get("Location"))
	})
}

func (suite *skinsystemTestSuite) TestSkinGET() {
	for _, testCase := range skinsTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/skins?name=mock_username", nil)
			w := httptest.NewRecorder()

			suite.App.Handler().ServeHTTP(w, req)

			testCase.AfterTest(suite, w.Result())
		})
	}

	suite.RunSubTest("Do not pass name param", func() {

		req := httptest.NewRequest("GET", "http://chrly/skins", nil)
		w := httptest.NewRecorder()

		suite.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		suite.Equal(400, resp.StatusCode)
	})
}

/************************
 * Get cape tests cases *
 ************************/

var capesTestsCases = []*skinsystemTestCase{
	{
		Name: "Username exists in the local storage",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(createCapeModel(), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			responseData, _ := ioutil.ReadAll(response.Body)
			suite.Equal(createCape(), responseData)
			suite.Equal("image/png", response.Header.Get("Content-Type"))
		},
	},
	{
		Name: "Username doesn't exists on the local storage, but exists on Mojang and has textures",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(nil, &CapeNotFoundError{Who: "mock_username"})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponse(true, true), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(301, response.StatusCode)
			suite.Equal("http://mojang/cape.png", response.Header.Get("Location"))
		},
	},
	{
		Name: "Username doesn't exists on the local storage, but exists on Mojang and has no textures",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(nil, &CapeNotFoundError{Who: "mock_username"})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponse(false, false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
	{
		Name: "Username doesn't exists on the local storage and doesn't exists on Mojang",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(nil, &CapeNotFoundError{Who: "mock_username"})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
}

func (suite *skinsystemTestSuite) TestCape() {
	for _, testCase := range capesTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/cloaks/mock_username", nil)
			w := httptest.NewRecorder()

			suite.App.Handler().ServeHTTP(w, req)

			testCase.AfterTest(suite, w.Result())
		})
	}

	suite.RunSubTest("Pass username with png extension", func() {
		suite.CapesRepository.On("FindByUsername", "mock_username").Return(createCapeModel(), nil)

		req := httptest.NewRequest("GET", "http://chrly/cloaks/mock_username.png", nil)
		w := httptest.NewRecorder()

		suite.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		suite.Equal(200, resp.StatusCode)
		responseData, _ := ioutil.ReadAll(resp.Body)
		suite.Equal(createCape(), responseData)
		suite.Equal("image/png", resp.Header.Get("Content-Type"))
	})
}

func (suite *skinsystemTestSuite) TestCapeGET() {
	for _, testCase := range capesTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/cloaks?name=mock_username", nil)
			w := httptest.NewRecorder()

			suite.App.Handler().ServeHTTP(w, req)

			testCase.AfterTest(suite, w.Result())
		})
	}

	suite.RunSubTest("Do not pass name param", func() {

		req := httptest.NewRequest("GET", "http://chrly/cloaks", nil)
		w := httptest.NewRecorder()

		suite.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		suite.Equal(400, resp.StatusCode)
	})
}

/****************************
 * Get textures tests cases *
 ****************************/

var texturesTestsCases = []*skinsystemTestCase{
	{
		Name: "Username exists and has skin, no cape",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(nil, &CapeNotFoundError{Who: "mock_username"})
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"SKIN": {
					"url": "http://chrly/skin.png"
				}
			}`, string(body))
		},
	},
	{
		Name: "Username exists and has slim skin, no cape",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(createSkinModel("mock_username", true), nil)
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(nil, &CapeNotFoundError{Who: "mock_username"})
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"SKIN": {
					"url": "http://chrly/skin.png",
					"metadata": {
						"model": "slim"
					}
				}
			}`, string(body))
		},
	},
	{
		Name: "Username exists and has cape, no skin",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(nil, &SkinNotFoundError{Who: "mock_username"})
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(createCapeModel(), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"CAPE": {
					"url": "http://chrly/cloaks/mock_username"
				}
			}`, string(body))
		},
	},
	{
		Name: "Username exists and has both skin and cape",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(createCapeModel(), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"SKIN": {
					"url": "http://chrly/skin.png"
				},
				"CAPE": {
					"url": "http://chrly/cloaks/mock_username"
				}
			}`, string(body))
		},
	},
	{
		Name: "Username not exists, but Mojang profile available",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(nil, &SkinNotFoundError{})
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(nil, &CapeNotFoundError{})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(createMojangResponse(true, true), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"SKIN": {
					"url": "http://mojang/skin.png"
				},
				"CAPE": {
					"url": "http://mojang/cape.png"
				}
			}`, string(body))
		},
	},
	{
		Name: "Username not exists and Mojang profile unavailable",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(nil, &SkinNotFoundError{})
			suite.CapesRepository.On("FindByUsername", "mock_username").Return(nil, &CapeNotFoundError{})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
		},
	},
}

func (suite *skinsystemTestSuite) TestTextures() {
	for _, testCase := range texturesTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
			w := httptest.NewRecorder()

			suite.App.Handler().ServeHTTP(w, req)

			testCase.AfterTest(suite, w.Result())
		})
	}
}

/***********************************
 * Get signed textures tests cases *
 ***********************************/

type signedTexturesTestCase struct {
	Name       string
	AllowProxy bool
	BeforeTest func(suite *skinsystemTestSuite)
	AfterTest  func(suite *skinsystemTestSuite, response *http.Response)
}

var signedTexturesTestsCases = []*signedTexturesTestCase{
	{
		Name:       "Username exists",
		AllowProxy: false,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(createSkinModel("mock_username", true), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"id": "0f657aa8bfbe415db7005750090d3af3",
				"name": "mock_username",
				"properties": [
					{
						"name": "textures",
						"signature": "mocked signature",
						"value": "mocked textures base64"
					},
					{
						"name": "chrly",
						"value": "how do you tame a horse in Minecraft?"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:       "Username not exists",
		AllowProxy: false,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(nil, &SkinNotFoundError{Who: "mock_username"})
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Equal("", string(body))
		},
	},
	{
		Name:       "Username exists, but has no signed textures",
		AllowProxy: false,
		BeforeTest: func(suite *skinsystemTestSuite) {
			skinModel := createSkinModel("mock_username", true)
			skinModel.MojangTextures = ""
			skinModel.MojangSignature = ""
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(skinModel, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Equal("", string(body))
		},
	},
	{
		Name:       "Username not exists, but Mojang profile is available and proxying is enabled",
		AllowProxy: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(nil, &SkinNotFoundError{Who: "mock_username"})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponse(true, false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"id": "00000000000000000000000000000000",
				"name": "mock_username",
				"properties": [
					{
						"name": "textures",
						"value": "eyJ0aW1lc3RhbXAiOjE1NTYzOTg1NzIsInByb2ZpbGVJZCI6IjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vbW9qYW5nL3NraW4ucG5nIn19fQ=="
					},
					{
						"name": "chrly",
						"value": "how do you tame a horse in Minecraft?"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:       "Username not exists, Mojang profile is unavailable too and proxying is enabled",
		AllowProxy: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindByUsername", "mock_username").Return(nil, &SkinNotFoundError{Who: "mock_username"})
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Equal("", string(body))
		},
	},
}

func (suite *skinsystemTestSuite) TestSignedTextures() {
	for _, testCase := range signedTexturesTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			var target string
			if testCase.AllowProxy {
				target = "http://chrly/textures/signed/mock_username?proxy=true"
			} else {
				target = "http://chrly/textures/signed/mock_username"
			}

			req := httptest.NewRequest("GET", target, nil)
			w := httptest.NewRecorder()

			suite.App.Handler().ServeHTTP(w, req)

			testCase.AfterTest(suite, w.Result())
		})
	}
}

/****************
 * Custom tests *
 ****************/

func TestParseUsername(t *testing.T) {
	assert := testify.New(t)
	assert.Equal("test", parseUsername("test.png"), "Function should trim .png at end")
	assert.Equal("test", parseUsername("test"), "Function should return string itself, if it not contains .png at end")
}

/*************
 * Utilities *
 *************/

func createSkinModel(username string, isSlim bool) *model.Skin {
	return &model.Skin{
		UserId:          1,
		Username:        username,
		Uuid:            "0f657aa8-bfbe-415d-b700-5750090d3af3", // Use non nil UUID to pass validation in api tests
		SkinId:          1,
		Url:             "http://chrly/skin.png",
		MojangTextures:  "mocked textures base64",
		MojangSignature: "mocked signature",
		IsSlim:          isSlim,
	}
}

func createCape() []byte {
	img := image.NewAlpha(image.Rect(0, 0, 64, 32))
	writer := &bytes.Buffer{}
	_ = png.Encode(writer, img)
	pngBytes, _ := ioutil.ReadAll(writer)

	return pngBytes
}

func createCapeModel() *model.Cape {
	return &model.Cape{File: bytes.NewReader(createCape())}
}

func createMojangResponse(includeSkin bool, includeCape bool) *mojang.SignedTexturesResponse {
	timeZone, _ := time.LoadLocation("Europe/Minsk")
	textures := &mojang.TexturesProp{
		Timestamp:   time.Date(2019, 4, 27, 23, 56, 12, 0, timeZone).Unix(),
		ProfileID:   "00000000000000000000000000000000",
		ProfileName: "mock_username",
		Textures:    &mojang.TexturesResponse{},
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
		Id:   "00000000000000000000000000000000",
		Name: "mock_username",
		Props: []*mojang.Property{
			{
				Name:  "textures",
				Value: mojang.EncodeTextures(textures),
			},
		},
	}

	return response
}
