package http

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/model"
)

/***************
 * Setup mocks *
 ***************/

type apiTestSuite struct {
	suite.Suite

	App *Api

	SkinsRepository *skinsRepositoryMock
}

/********************
 * Setup test suite *
 ********************/

func (suite *apiTestSuite) SetupTest() {
	suite.SkinsRepository = &skinsRepositoryMock{}

	suite.App = &Api{
		SkinsRepo: suite.SkinsRepository,
	}
}

func (suite *apiTestSuite) TearDownTest() {
	suite.SkinsRepository.AssertExpectations(suite.T())
}

func (suite *apiTestSuite) RunSubTest(name string, subTest func()) {
	suite.SetupTest()
	suite.Run(name, subTest)
	suite.TearDownTest()
}

/*************
 * Run tests *
 *************/

func TestApi(t *testing.T) {
	suite.Run(t, new(apiTestSuite))
}

/*************************
 * Post skin tests cases *
 *************************/

type postSkinTestCase struct {
	Name       string
	Form       io.Reader
	BeforeTest func(suite *apiTestSuite)
	PanicErr   string
	AfterTest  func(suite *apiTestSuite, response *http.Response)
}

var postSkinTestsCases = []*postSkinTestCase{
	{
		Name: "Upload new identity with textures data",
		Form: bytes.NewBufferString(url.Values{
			"identityId": {"1"},
			"username":   {"mock_username"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"is1_8":      {"0"},
			"isSlim":     {"0"},
			"url":        {"http://example.com/skin.png"},
		}.Encode()),
		BeforeTest: func(suite *apiTestSuite) {
			suite.SkinsRepository.On("FindSkinByUserId", 1).Return(nil, nil)
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)
			suite.SkinsRepository.On("SaveSkin", mock.MatchedBy(func(model *model.Skin) bool {
				suite.Equal(1, model.UserId)
				suite.Equal("mock_username", model.Username)
				suite.Equal("0f657aa8-bfbe-415d-b700-5750090d3af3", model.Uuid)
				suite.Equal(5, model.SkinId)
				suite.False(model.Is1_8)
				suite.False(model.IsSlim)
				suite.Equal("http://example.com/skin.png", model.Url)

				return true
			})).Times(1).Return(nil)
		},
		AfterTest: func(suite *apiTestSuite, response *http.Response) {
			suite.Equal(201, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Empty(body)
		},
	},
	{
		Name: "Update exists identity by changing only textures data",
		Form: bytes.NewBufferString(url.Values{
			"identityId": {"1"},
			"username":   {"mock_username"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"is1_8":      {"1"},
			"isSlim":     {"1"},
			"url":        {"http://textures-server.com/skin.png"},
		}.Encode()),
		BeforeTest: func(suite *apiTestSuite) {
			suite.SkinsRepository.On("FindSkinByUserId", 1).Return(createSkinModel("mock_username", false), nil)
			suite.SkinsRepository.On("SaveSkin", mock.MatchedBy(func(model *model.Skin) bool {
				suite.Equal(1, model.UserId)
				suite.Equal("mock_username", model.Username)
				suite.Equal("0f657aa8-bfbe-415d-b700-5750090d3af3", model.Uuid)
				suite.Equal(5, model.SkinId)
				suite.True(model.Is1_8)
				suite.True(model.IsSlim)
				suite.Equal("http://textures-server.com/skin.png", model.Url)

				return true
			})).Times(1).Return(nil)
		},
		AfterTest: func(suite *apiTestSuite, response *http.Response) {
			suite.Equal(201, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Empty(body)
		},
	},
	{
		Name: "Update exists identity by changing textures data to empty",
		Form: bytes.NewBufferString(url.Values{
			"identityId":      {"1"},
			"username":        {"mock_username"},
			"uuid":            {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":          {"0"},
			"is1_8":           {"0"},
			"isSlim":          {"0"},
			"url":             {""},
			"mojangTextures":  {""},
			"mojangSignature": {""},
		}.Encode()),
		BeforeTest: func(suite *apiTestSuite) {
			suite.SkinsRepository.On("FindSkinByUserId", 1).Return(createSkinModel("mock_username", false), nil)
			suite.SkinsRepository.On("SaveSkin", mock.MatchedBy(func(model *model.Skin) bool {
				suite.Equal(1, model.UserId)
				suite.Equal("mock_username", model.Username)
				suite.Equal("0f657aa8-bfbe-415d-b700-5750090d3af3", model.Uuid)
				suite.Equal(0, model.SkinId)
				suite.False(model.Is1_8)
				suite.False(model.IsSlim)
				suite.Equal("", model.Url)
				suite.Equal("", model.MojangTextures)
				suite.Equal("", model.MojangSignature)

				return true
			})).Times(1).Return(nil)
		},
		AfterTest: func(suite *apiTestSuite, response *http.Response) {
			suite.Equal(201, response.StatusCode)
			body, _ := io.ReadAll(response.Body)
			suite.Equal("", string(body))
		},
	},
	{
		Name: "Update exists identity by changing its identityId",
		Form: bytes.NewBufferString(url.Values{
			"identityId": {"2"},
			"username":   {"mock_username"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"is1_8":      {"0"},
			"isSlim":     {"0"},
			"url":        {"http://example.com/skin.png"},
		}.Encode()),
		BeforeTest: func(suite *apiTestSuite) {
			suite.SkinsRepository.On("FindSkinByUserId", 2).Return(nil, nil)
			suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
			suite.SkinsRepository.On("RemoveSkinByUsername", "mock_username").Times(1).Return(nil)
			suite.SkinsRepository.On("SaveSkin", mock.MatchedBy(func(model *model.Skin) bool {
				suite.Equal(2, model.UserId)
				suite.Equal("mock_username", model.Username)
				suite.Equal("0f657aa8-bfbe-415d-b700-5750090d3af3", model.Uuid)

				return true
			})).Times(1).Return(nil)
		},
		AfterTest: func(suite *apiTestSuite, response *http.Response) {
			suite.Equal(201, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Empty(body)
		},
	},
	{
		Name: "Update exists identity by changing its username",
		Form: bytes.NewBufferString(url.Values{
			"identityId": {"1"},
			"username":   {"changed_username"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"is1_8":      {"0"},
			"isSlim":     {"0"},
			"url":        {"http://example.com/skin.png"},
		}.Encode()),
		BeforeTest: func(suite *apiTestSuite) {
			suite.SkinsRepository.On("FindSkinByUserId", 1).Return(createSkinModel("mock_username", false), nil)
			suite.SkinsRepository.On("RemoveSkinByUserId", 1).Times(1).Return(nil)
			suite.SkinsRepository.On("SaveSkin", mock.MatchedBy(func(model *model.Skin) bool {
				suite.Equal(1, model.UserId)
				suite.Equal("changed_username", model.Username)
				suite.Equal("0f657aa8-bfbe-415d-b700-5750090d3af3", model.Uuid)

				return true
			})).Times(1).Return(nil)
		},
		AfterTest: func(suite *apiTestSuite, response *http.Response) {
			suite.Equal(201, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Empty(body)
		},
	},
	{
		Name: "Handle an error when loading the data from the repository",
		Form: bytes.NewBufferString(url.Values{
			"identityId": {"1"},
			"username":   {"changed_username"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"is1_8":      {"0"},
			"isSlim":     {"0"},
			"url":        {"http://example.com/skin.png"},
		}.Encode()),
		BeforeTest: func(suite *apiTestSuite) {
			suite.SkinsRepository.On("FindSkinByUserId", 1).Return(nil, errors.New("can't find skin by user id"))
		},
		PanicErr: "can't find skin by user id",
	},
	{
		Name: "Handle an error when saving the data into the repository",
		Form: bytes.NewBufferString(url.Values{
			"identityId": {"1"},
			"username":   {"mock_username"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"is1_8":      {"1"},
			"isSlim":     {"1"},
			"url":        {"http://textures-server.com/skin.png"},
		}.Encode()),
		BeforeTest: func(suite *apiTestSuite) {
			suite.SkinsRepository.On("FindSkinByUserId", 1).Return(createSkinModel("mock_username", false), nil)
			suite.SkinsRepository.On("SaveSkin", mock.Anything).Return(errors.New("can't save textures"))
		},
		PanicErr: "can't save textures",
	},
}

func (suite *apiTestSuite) TestPostSkin() {
	for _, testCase := range postSkinTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("POST", "http://chrly/skins", testCase.Form)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
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

	suite.RunSubTest("Get errors about required fields", func() {
		req := httptest.NewRequest("POST", "http://chrly/skins", bytes.NewBufferString(url.Values{
			"mojangTextures": {"someBase64EncodedString"},
		}.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		suite.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		suite.Equal(400, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		suite.JSONEq(`{
			"errors": {
				"identityId": [
					"The identityId field is required",
					"The identityId field must be numeric",
					"The identityId field must be minimum 1 char"
				],
				"skinId": [
					"The skinId field is required",
					"The skinId field must be numeric",
					"The skinId field must be numeric value between 0 and 0"
				],
				"username": [
					"The username field is required"
				],
				"uuid": [
					"The uuid field is required",
					"The uuid field must contain valid UUID"
				],
				"mojangSignature": [
					"The mojangSignature field is required"
				]
			}
		}`, string(body))
	})
}

/**************************************
 * Delete skin by user id tests cases *
 **************************************/

func (suite *apiTestSuite) TestDeleteByUserId() {
	suite.RunSubTest("Delete skin by its identity id", func() {
		suite.SkinsRepository.On("FindSkinByUserId", 1).Return(createSkinModel("mock_username", false), nil)
		suite.SkinsRepository.On("RemoveSkinByUserId", 1).Once().Return(nil)

		req := httptest.NewRequest("DELETE", "http://chrly/skins/id:1", nil)
		w := httptest.NewRecorder()

		suite.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		suite.Equal(204, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		suite.Empty(body)
	})

	suite.RunSubTest("Try to remove not exists identity id", func() {
		suite.SkinsRepository.On("FindSkinByUserId", 1).Return(nil, nil)

		req := httptest.NewRequest("DELETE", "http://chrly/skins/id:1", nil)
		w := httptest.NewRecorder()

		suite.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		suite.Equal(404, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		suite.JSONEq(`[
			"Cannot find record for the requested identifier"
		]`, string(body))
	})
}

/***************************************
 * Delete skin by username tests cases *
 ***************************************/

func (suite *apiTestSuite) TestDeleteByUsername() {
	suite.RunSubTest("Delete skin by its identity username", func() {
		suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(createSkinModel("mock_username", false), nil)
		suite.SkinsRepository.On("RemoveSkinByUserId", 1).Once().Return(nil)

		req := httptest.NewRequest("DELETE", "http://chrly/skins/mock_username", nil)
		w := httptest.NewRecorder()

		suite.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		suite.Equal(204, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		suite.Empty(body)
	})

	suite.RunSubTest("Try to remove not exists identity username", func() {
		suite.SkinsRepository.On("FindSkinByUsername", "mock_username").Return(nil, nil)

		req := httptest.NewRequest("DELETE", "http://chrly/skins/mock_username", nil)
		w := httptest.NewRecorder()

		suite.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		suite.Equal(404, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		suite.JSONEq(`[
			"Cannot find record for the requested identifier"
		]`, string(body))
	})
}

/*************
 * Utilities *
 *************/

// base64 https://github.com/mathiasbynens/small/blob/0ca3c51/png-transparent.png
var OnePxPng = []byte("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQottAAAAABJRU5ErkJggg==")

func loadSkinFile() []byte {
	result := make([]byte, 92)
	_, err := base64.StdEncoding.Decode(result, OnePxPng)
	if err != nil {
		panic(err)
	}

	return result
}
