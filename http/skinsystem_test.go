package http

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func (m *skinsRepositoryMock) FindSkinByUsername(username string) (*model.Skin, error) {
	args := m.Called(username)
	var result *model.Skin
	if casted, ok := args.Get(0).(*model.Skin); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *skinsRepositoryMock) FindSkinByUserId(id int) (*model.Skin, error) {
	args := m.Called(id)
	var result *model.Skin
	if casted, ok := args.Get(0).(*model.Skin); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *skinsRepositoryMock) SaveSkin(skin *model.Skin) error {
	args := m.Called(skin)
	return args.Error(0)
}

func (m *skinsRepositoryMock) RemoveSkinByUserId(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *skinsRepositoryMock) RemoveSkinByUsername(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

type capesRepositoryMock struct {
	mock.Mock
}

func (m *capesRepositoryMock) FindCapeByUsername(username string) (*model.Cape, error) {
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

type texturesSignerMock struct {
	mock.Mock
}

func (m *texturesSignerMock) SignTextures(textures string) (string, error) {
	args := m.Called(textures)
	return args.String(0), args.Error(1)
}

func (m *texturesSignerMock) GetPublicKey() (*rsa.PublicKey, error) {
	args := m.Called()
	var publicKey *rsa.PublicKey
	if casted, ok := args.Get(0).(*rsa.PublicKey); ok {
		publicKey = casted
	}

	return publicKey, args.Error(1)
}

type skinsystemTestSuite struct {
	suite.Suite

	App *Skinsystem

	SkinsRepository        *skinsRepositoryMock
	CapesRepository        *capesRepositoryMock
	MojangTexturesProvider *mojangTexturesProviderMock
	TexturesSigner         *texturesSignerMock
	Emitter                *emitterMock
}

/********************
 * Setup test suite *
 ********************/

func (suite *skinsystemTestSuite) SetupTest() {
	timeNow = func() time.Time {
		CET, _ := time.LoadLocation("CET")
		return time.Date(2021, 02, 25, 01, 50, 23, 0, CET)
	}

	suite.SkinsRepository = &skinsRepositoryMock{}
	suite.CapesRepository = &capesRepositoryMock{}
	suite.MojangTexturesProvider = &mojangTexturesProviderMock{}
	suite.TexturesSigner = &texturesSignerMock{}
	suite.Emitter = &emitterMock{}

	suite.TexturesSigner.On("SignTextures", "texturesParamValue").Times(1).Return("texturesParamSignature", nil)

	suite.App, _ = NewSkinsystem(
		suite.Emitter,
		suite.SkinsRepository,
		suite.CapesRepository,
		suite.MojangTexturesProvider,
		suite.TexturesSigner,
		"texturesParamName",
		"texturesParamValue",
	)
}

func (suite *skinsystemTestSuite) TearDownTest() {
	suite.SkinsRepository.AssertExpectations(suite.T())
	suite.CapesRepository.AssertExpectations(suite.T())
	suite.MojangTexturesProvider.AssertExpectations(suite.T())
	suite.TexturesSigner.AssertExpectations(suite.T())
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
	PanicErr   string
	AfterTest  func(suite *skinsystemTestSuite, response *http.Response)
}

/************************
 * Get skin tests cases *
 ************************/

var skinsTestsCases = []*skinsystemTestCase{
	{
		Name: "Username exists in the local storage",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(301, response.StatusCode)
			suite.Equal("http://chrly/skin.png", response.Header.Get("Location"))
		},
	},
	{
		Name: "Username doesn't exists on the local storage, but exists on Mojang and has textures",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponseWithTextures(true, false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(301, response.StatusCode)
			suite.Equal("http://mojang/skin.png", response.Header.Get("Location"))
		},
	},
	{
		Name: "Username doesn't exists on the local storage, but exists on Mojang and has no skin texture",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponseWithTextures(false, false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
	{
		Name: "Username doesn't exists on the local storage, but exists on Mojang and has an empty properties",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createEmptyMojangResponse(), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
	{
		Name: "Username doesn't exists on the local storage and doesn't exists on Mojang",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
	{
		Name: "Receive an error from the SkinsRepository",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, errors.New("skins repository error"))
		},
		PanicErr: "skins repository error",
	},
}

func (suite *skinsystemTestSuite) TestSkin() {
	for _, testCase := range skinsTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/skins/mock_username", nil)
			w := httptest.NewRecorder()

			if testCase.PanicErr != "" {
				suite.PanicsWithError(testCase.PanicErr, func() {
					suite.App.Handler().ServeHTTP(w, req)
				})
			} else {
				suite.App.Handler().ServeHTTP(w, req)
				testCase.AfterTest(suite, w.Result())
			}
		})
	}

	suite.RunSubTest("Pass username with png extension", func() {
		suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
		suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(nil, nil)

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

			if testCase.PanicErr != "" {
				suite.PanicsWithError(testCase.PanicErr, func() {
					suite.App.Handler().ServeHTTP(w, req)
				})
			} else {
				suite.App.Handler().ServeHTTP(w, req)
				testCase.AfterTest(suite, w.Result())
			}
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
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(createCapeModel(), nil)
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
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponseWithTextures(true, true), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(301, response.StatusCode)
			suite.Equal("http://mojang/cape.png", response.Header.Get("Location"))
		},
	},
	{
		Name: "Username doesn't exists on the local storage, but exists on Mojang and has no cape texture",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponseWithTextures(false, false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
	{
		Name: "Username doesn't exists on the local storage, but exists on Mojang and has an empty properties",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createEmptyMojangResponse(), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
	{
		Name: "Username doesn't exists on the local storage and doesn't exists on Mojang",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(404, response.StatusCode)
		},
	},
	{
		Name: "Receive an error from the SkinsRepository",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, errors.New("skins repository error"))
		},
		PanicErr: "skins repository error",
	},
}

func (suite *skinsystemTestSuite) TestCape() {
	for _, testCase := range capesTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/cloaks/mock_username", nil)
			w := httptest.NewRecorder()

			if testCase.PanicErr != "" {
				suite.PanicsWithError(testCase.PanicErr, func() {
					suite.App.Handler().ServeHTTP(w, req)
				})
			} else {
				suite.App.Handler().ServeHTTP(w, req)
				testCase.AfterTest(suite, w.Result())
			}
		})
	}

	suite.RunSubTest("Pass username with png extension", func() {
		suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
		suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(createCapeModel(), nil)

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

			if testCase.PanicErr != "" {
				suite.PanicsWithError(testCase.PanicErr, func() {
					suite.App.Handler().ServeHTTP(w, req)
				})
			} else {
				suite.App.Handler().ServeHTTP(w, req)
				testCase.AfterTest(suite, w.Result())
			}
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
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(nil, nil)
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
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", true), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(nil, nil)
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
	// There is no case when the user has cape, but has no skin.
	// In v5 we will rework textures repositories to be more generic about source of textures,
	// but right now it's not possible to return profile entity with a cape only.
	{
		Name: "Username exists and has both skin and cape",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(createCapeModel(), nil)
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
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(createMojangResponseWithTextures(true, true), nil)
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
		Name: "Username not exists, but Mojang profile available, but there is an empty skin and cape textures",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(createMojangResponseWithTextures(false, false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
		},
	},
	{
		Name: "Username not exists, but Mojang profile available, but there is an empty properties",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(createEmptyMojangResponse(), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
		},
	},
	{
		Name: "Username not exists and Mojang profile unavailable",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Equal("", string(body))
		},
	},
	{
		Name: "Receive an error from the SkinsRepository",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, errors.New("skins repository error"))
		},
		PanicErr: "skins repository error",
	},
}

func (suite *skinsystemTestSuite) TestTextures() {
	for _, testCase := range texturesTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
			w := httptest.NewRecorder()

			if testCase.PanicErr != "" {
				suite.PanicsWithError(testCase.PanicErr, func() {
					suite.App.Handler().ServeHTTP(w, req)
				})
			} else {
				suite.App.Handler().ServeHTTP(w, req)
				testCase.AfterTest(suite, w.Result())
			}
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
	PanicErr   string
	AfterTest  func(suite *skinsystemTestSuite, response *http.Response)
}

var signedTexturesTestsCases = []*signedTexturesTestCase{
	{
		Name:       "Username exists",
		AllowProxy: false,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", true), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(nil, nil)
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
						"name": "texturesParamName",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:       "Username not exists",
		AllowProxy: false,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
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
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(skinModel, nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(nil, nil)
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
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(createMojangResponseWithTextures(true, false), nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"id": "292a1db7353d476ca99cab8f57mojang",
				"name": "mock_username",
				"properties": [
					{
						"name": "textures",
						"value": "eyJ0aW1lc3RhbXAiOjE1NTYzOTg1NzIwMDAsInByb2ZpbGVJZCI6IjI5MmExZGI3MzUzZDQ3NmNhOTljYWI4ZjU3bW9qYW5nIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vbW9qYW5nL3NraW4ucG5nIn19fQ==",
						"signature": "mojang signature"
					},
					{
						"name": "texturesParamName",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:       "Username not exists, Mojang profile is unavailable too and proxying is enabled",
		AllowProxy: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Equal("", string(body))
		},
	},
	{
		Name: "Receive an error from the SkinsRepository",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, errors.New("skins repository error"))
		},
		PanicErr: "skins repository error",
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

			if testCase.PanicErr != "" {
				suite.PanicsWithError(testCase.PanicErr, func() {
					suite.App.Handler().ServeHTTP(w, req)
				})
			} else {
				suite.App.Handler().ServeHTTP(w, req)
				testCase.AfterTest(suite, w.Result())
			}
		})
	}
}

/***************************
 * Get profile tests cases *
 ***************************/

type profileTestCase struct {
	Name          string
	Signed        bool
	ForceResponse string
	BeforeTest    func(suite *skinsystemTestSuite)
	PanicErr      string
	AfterTest     func(suite *skinsystemTestSuite, response *http.Response)
}

var profileTestsCases = []*profileTestCase{
	{
		Name: "Username exists and has both skin and cape, don't sign",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(createCapeModel(), nil)
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
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjBmNjU3YWE4YmZiZTQxNWRiNzAwNTc1MDA5MGQzYWYzIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vY2hybHkvc2tpbi5wbmcifSwiQ0FQRSI6eyJ1cmwiOiJodHRwOi8vY2hybHkvY2xvYWtzL21vY2tfdXNlcm5hbWUifX19"
					},
					{
						"name": "texturesParamName",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:   "Username exists and has both skin and cape",
		Signed: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(createCapeModel(), nil)
			suite.TexturesSigner.On("SignTextures", "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjBmNjU3YWE4YmZiZTQxNWRiNzAwNTc1MDA5MGQzYWYzIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vY2hybHkvc2tpbi5wbmcifSwiQ0FQRSI6eyJ1cmwiOiJodHRwOi8vY2hybHkvY2xvYWtzL21vY2tfdXNlcm5hbWUifX19").Return("textures signature", nil)
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
						"signature": "textures signature",
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjBmNjU3YWE4YmZiZTQxNWRiNzAwNTc1MDA5MGQzYWYzIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vY2hybHkvc2tpbi5wbmcifSwiQ0FQRSI6eyJ1cmwiOiJodHRwOi8vY2hybHkvY2xvYWtzL21vY2tfdXNlcm5hbWUifX19"
					},
					{
						"name": "texturesParamName",
						"signature": "texturesParamSignature",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:   "Username exists and has skin, no cape",
		Signed: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(nil, nil)
			suite.TexturesSigner.On("SignTextures", mock.Anything).Return("textures signature", nil)
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
						"signature": "textures signature",
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjBmNjU3YWE4YmZiZTQxNWRiNzAwNTc1MDA5MGQzYWYzIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vY2hybHkvc2tpbi5wbmcifX19"
					},
					{
						"name": "texturesParamName",
						"signature": "texturesParamSignature",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:   "Username exists and has slim skin, no cape",
		Signed: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", true), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(nil, nil)
			suite.TexturesSigner.On("SignTextures", mock.Anything).Return("textures signature", nil)
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
						"signature": "textures signature",
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjBmNjU3YWE4YmZiZTQxNWRiNzAwNTc1MDA5MGQzYWYzIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vY2hybHkvc2tpbi5wbmciLCJtZXRhZGF0YSI6eyJtb2RlbCI6InNsaW0ifX19fQ=="
					},
					{
						"name": "texturesParamName",
						"signature": "texturesParamSignature",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:   "Username exists, but has no skin and Mojang profile with textures available",
		Signed: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			skin := createSkinModel("mock_username", false)
			skin.SkinId = 0
			skin.Url = ""

			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(skin, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(createMojangResponseWithTextures(true, true), nil)
			suite.TexturesSigner.On("SignTextures", mock.Anything).Return("chrly signature", nil)
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
						"signature": "chrly signature",
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjBmNjU3YWE4YmZiZTQxNWRiNzAwNTc1MDA5MGQzYWYzIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vbW9qYW5nL3NraW4ucG5nIn0sIkNBUEUiOnsidXJsIjoiaHR0cDovL21vamFuZy9jYXBlLnBuZyJ9fX0="
					},
					{
						"name": "texturesParamName",
						"signature": "texturesParamSignature",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:   "Username exists, but has no skin and Mojang textures proxy returned an error",
		Signed: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			skin := createSkinModel("mock_username", false)
			skin.SkinId = 0
			skin.Url = ""

			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(skin, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(nil, errors.New("shit happened"))
			suite.TexturesSigner.On("SignTextures", mock.Anything).Return("chrly signature", nil)
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
						"signature": "chrly signature",
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjBmNjU3YWE4YmZiZTQxNWRiNzAwNTc1MDA5MGQzYWYzIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnt9fQ=="
					},
					{
						"name": "texturesParamName",
						"signature": "texturesParamSignature",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:   "Username not exists, but Mojang profile with textures available",
		Signed: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(createMojangResponseWithTextures(true, true), nil)
			suite.TexturesSigner.On("SignTextures", mock.Anything).Return("chrly signature", nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"id": "292a1db7353d476ca99cab8f57mojang",
				"name": "mock_username",
				"properties": [
					{
						"name": "textures",
						"signature": "chrly signature",
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjI5MmExZGI3MzUzZDQ3NmNhOTljYWI4ZjU3bW9qYW5nIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vbW9qYW5nL3NraW4ucG5nIn0sIkNBUEUiOnsidXJsIjoiaHR0cDovL21vamFuZy9jYXBlLnBuZyJ9fX0="
					},
					{
						"name": "texturesParamName",
						"signature": "texturesParamSignature",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:   "Username not exists, but Mojang profile available, but there is an empty skin and cape textures",
		Signed: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(createMojangResponseWithTextures(false, false), nil)
			suite.TexturesSigner.On("SignTextures", mock.Anything).Return("chrly signature", nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"id": "292a1db7353d476ca99cab8f57mojang",
				"name": "mock_username",
				"properties": [
					{
						"name": "textures",
						"signature": "chrly signature",
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjI5MmExZGI3MzUzZDQ3NmNhOTljYWI4ZjU3bW9qYW5nIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnt9fQ=="
					},
					{
						"name": "texturesParamName",
						"signature": "texturesParamSignature",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name:   "Username not exists, but Mojang profile available, but there is an empty properties",
		Signed: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(createEmptyMojangResponse(), nil)
			suite.TexturesSigner.On("SignTextures", mock.Anything).Return("chrly signature", nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"id": "292a1db7353d476ca99cab8f57mojang",
				"name": "mock_username",
				"properties": [
					{
						"name": "textures",
						"signature": "chrly signature",
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6IjI5MmExZGI3MzUzZDQ3NmNhOTljYWI4ZjU3bW9qYW5nIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnt9fQ=="
					},
					{
						"name": "texturesParamName",
						"signature": "texturesParamSignature",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name: "Username not exists and Mojang profile unavailable",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(nil, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Equal("", string(body))
		},
	},
	{
		Name:          "Username not exists and Mojang profile unavailable, but there is a forceResponse param",
		ForceResponse: "a12e41a4-e8e5-4503-987e-0adacf72ab93",
		Signed:        true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(nil, nil)
			suite.TexturesSigner.On("SignTextures", mock.Anything).Return("chrly signature", nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"id": "a12e41a4e8e54503987e0adacf72ab93",
				"name": "mock_username",
				"properties": [
					{
						"name": "textures",
						"signature": "chrly signature",
						"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6ImExMmU0MWE0ZThlNTQ1MDM5ODdlMGFkYWNmNzJhYjkzIiwicHJvZmlsZU5hbWUiOiJtb2NrX3VzZXJuYW1lIiwidGV4dHVyZXMiOnt9fQ=="
					},
					{
						"name": "texturesParamName",
						"signature": "texturesParamSignature",
						"value": "texturesParamValue"
					}
				]
			}`, string(body))
		},
	},
	{
		Name: "Username not exists and Mojang textures proxy returned an error",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.MojangTexturesProvider.On("GetForUsername", "mock_username").Once().Return(nil, errors.New("mojang textures provider error"))
		},
		PanicErr: "mojang textures provider error",
	},
	{
		Name: "Receive an error from the SkinsRepository",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, errors.New("skins repository error"))
		},
		PanicErr: "skins repository error",
	},
	{
		Name:   "Receive an error from the TexturesSigner",
		Signed: true,
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.CapesRepository.On("FindCapeByUsername", "mock_username").Return(nil, nil)
			suite.TexturesSigner.On("SignTextures", mock.Anything).Return("", errors.New("textures signer error"))
		},
		PanicErr: "textures signer error",
	},
}

func (suite *skinsystemTestSuite) TestProfile() {
	for _, testCase := range profileTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			u, _ := url.Parse("http://chrly/profile/mock_username")
			q := make(url.Values)
			if testCase.Signed {
				q.Set("unsigned", "false")
			}

			if testCase.ForceResponse != "" {
				q.Set("onUnknownProfileRespondWithUuid", testCase.ForceResponse)
			}

			u.RawQuery = q.Encode()
			req := httptest.NewRequest("GET", u.String(), nil)
			w := httptest.NewRecorder()

			if testCase.PanicErr != "" {
				suite.PanicsWithError(testCase.PanicErr, func() {
					suite.App.Handler().ServeHTTP(w, req)
				})
			} else {
				suite.App.Handler().ServeHTTP(w, req)
				testCase.AfterTest(suite, w.Result())
			}
		})
	}
}

/***************************
 * Get profile tests cases *
 ***************************/

type signingKeyTestCase struct {
	Name       string
	KeyFormat  string
	BeforeTest func(suite *skinsystemTestSuite)
	PanicErr   string
	AfterTest  func(suite *skinsystemTestSuite, response *http.Response)
}

var signingKeyTestsCases = []*signingKeyTestCase{
	{
		Name:      "Get public key in DER format",
		KeyFormat: "DER",
		BeforeTest: func(suite *skinsystemTestSuite) {
			pubPem, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANbUpVCZkMKpfvYZ08W3lumdAaYxLBnm\nUDlzHBQH3DpYef5WCO32TDU6feIJ58A0lAywgtZ4wwi2dGHOz/1hAvcCAwEAAQ==\n-----END PUBLIC KEY-----"))
			publicKey, _ := x509.ParsePKIXPublicKey(pubPem.Bytes)

			suite.TexturesSigner.On("GetPublicKey").Return(publicKey, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/octet-stream", response.Header.Get("Content-Type"))
			suite.Equal("attachment; filename=\"yggdrasil_session_pubkey.der\"", response.Header.Get("Content-Disposition"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.Equal([]byte{48, 92, 48, 13, 6, 9, 42, 134, 72, 134, 247, 13, 1, 1, 1, 5, 0, 3, 75, 0, 48, 72, 2, 65, 0, 214, 212, 165, 80, 153, 144, 194, 169, 126, 246, 25, 211, 197, 183, 150, 233, 157, 1, 166, 49, 44, 25, 230, 80, 57, 115, 28, 20, 7, 220, 58, 88, 121, 254, 86, 8, 237, 246, 76, 53, 58, 125, 226, 9, 231, 192, 52, 148, 12, 176, 130, 214, 120, 195, 8, 182, 116, 97, 206, 207, 253, 97, 2, 247, 2, 3, 1, 0, 1}, body)
		},
	},
	{
		Name:      "Get public key in PEM format",
		KeyFormat: "PEM",
		BeforeTest: func(suite *skinsystemTestSuite) {
			pubPem, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANbUpVCZkMKpfvYZ08W3lumdAaYxLBnm\nUDlzHBQH3DpYef5WCO32TDU6feIJ58A0lAywgtZ4wwi2dGHOz/1hAvcCAwEAAQ==\n-----END PUBLIC KEY-----"))
			publicKey, _ := x509.ParsePKIXPublicKey(pubPem.Bytes)

			suite.TexturesSigner.On("GetPublicKey").Return(publicKey, nil)
		},
		AfterTest: func(suite *skinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("text/plain; charset=utf-8", response.Header.Get("Content-Type"))
			suite.Equal("attachment; filename=\"yggdrasil_session_pubkey.pem\"", response.Header.Get("Content-Disposition"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.Equal("-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANbUpVCZkMKpfvYZ08W3lumdAaYxLBnm\nUDlzHBQH3DpYef5WCO32TDU6feIJ58A0lAywgtZ4wwi2dGHOz/1hAvcCAwEAAQ==\n-----END PUBLIC KEY-----\n", string(body))
		},
	},
	{
		Name:      "Error while obtaining public key",
		KeyFormat: "DER",
		BeforeTest: func(suite *skinsystemTestSuite) {
			suite.TexturesSigner.On("GetPublicKey").Return(nil, errors.New("textures signer error"))
		},
		PanicErr: "textures signer error",
	},
}

func (suite *skinsystemTestSuite) TestSignatureVerificationKey() {
	for _, testCase := range signingKeyTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/signature-verification-key."+strings.ToLower(testCase.KeyFormat), nil)
			w := httptest.NewRecorder()

			if testCase.PanicErr != "" {
				suite.PanicsWithError(testCase.PanicErr, func() {
					suite.App.Handler().ServeHTTP(w, req)
				})
			} else {
				suite.App.Handler().ServeHTTP(w, req)
				testCase.AfterTest(suite, w.Result())
			}
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

func createEmptyMojangResponse() *mojang.SignedTexturesResponse {
	return &mojang.SignedTexturesResponse{
		Id:    "292a1db7353d476ca99cab8f57mojang",
		Name:  "mock_username",
		Props: []*mojang.Property{},
	}
}

func createMojangResponseWithTextures(includeSkin bool, includeCape bool) *mojang.SignedTexturesResponse {
	timeZone, _ := time.LoadLocation("Europe/Minsk")
	textures := &mojang.TexturesProp{
		Timestamp:   time.Date(2019, 4, 27, 23, 56, 12, 0, timeZone).UnixNano() / int64(time.Millisecond),
		ProfileID:   "292a1db7353d476ca99cab8f57mojang",
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

	response := createEmptyMojangResponse()
	response.Props = append(response.Props, &mojang.Property{
		Name:      "textures",
		Value:     mojang.EncodeTextures(textures),
		Signature: "mojang signature",
	})

	return response
}
