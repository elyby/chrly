package http

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/api/mojang"
)

/***************
 * Setup mocks *
 ***************/

type uuidsProviderMock struct {
	mock.Mock
}

func (m *uuidsProviderMock) GetUuid(username string) (*mojang.ProfileInfo, error) {
	args := m.Called(username)
	var result *mojang.ProfileInfo
	if casted, ok := args.Get(0).(*mojang.ProfileInfo); ok {
		result = casted
	}

	return result, args.Error(1)
}

type uuidsWorkerTestSuite struct {
	suite.Suite

	App *UUIDsWorker

	UuidsProvider *uuidsProviderMock
	Emitter       *emitterMock
}

/********************
 * Setup test suite *
 ********************/

func (suite *uuidsWorkerTestSuite) SetupTest() {
	suite.UuidsProvider = &uuidsProviderMock{}
	suite.Emitter = &emitterMock{}

	suite.App = &UUIDsWorker{
		UUIDsProvider: suite.UuidsProvider,
		Emitter:       suite.Emitter,
	}
}

func (suite *uuidsWorkerTestSuite) TearDownTest() {
	suite.UuidsProvider.AssertExpectations(suite.T())
	suite.Emitter.AssertExpectations(suite.T())
}

func (suite *uuidsWorkerTestSuite) RunSubTest(name string, subTest func()) {
	suite.SetupTest()
	suite.Run(name, subTest)
	suite.TearDownTest()
}

/*************
 * Run tests *
 *************/

func TestUUIDsWorker(t *testing.T) {
	suite.Run(t, new(uuidsWorkerTestSuite))
}

type uuidsWorkerTestCase struct {
	Name       string
	BeforeTest func(suite *uuidsWorkerTestSuite)
	AfterTest  func(suite *uuidsWorkerTestSuite, response *http.Response)
}

/************************
 * Get UUID tests cases *
 ************************/

var getUuidTestsCases = []*uuidsWorkerTestCase{
	{
		Name: "Success provider response",
		BeforeTest: func(suite *uuidsWorkerTestSuite) {
			suite.Emitter.On("Emit", "skinsystem:before_request", mock.Anything)
			suite.Emitter.On("Emit", "skinsystem:after_request", mock.Anything, 200)
			suite.UuidsProvider.On("GetUuid", "mock_username").Return(&mojang.ProfileInfo{
				Id:   "0fcc38620f1845f3a54e1b523c1bd1c7",
				Name: "mock_username",
			}, nil)
		},
		AfterTest: func(suite *uuidsWorkerTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"id": "0fcc38620f1845f3a54e1b523c1bd1c7",
				"name": "mock_username"
			}`, string(body))
		},
	},
	{
		Name: "Receive empty response from UUIDs provider",
		BeforeTest: func(suite *uuidsWorkerTestSuite) {
			suite.Emitter.On("Emit", "skinsystem:before_request", mock.Anything)
			suite.Emitter.On("Emit", "skinsystem:after_request", mock.Anything, 204)
			suite.UuidsProvider.On("GetUuid", "mock_username").Return(nil, nil)
		},
		AfterTest: func(suite *uuidsWorkerTestSuite, response *http.Response) {
			suite.Equal(204, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Assert().Empty(body)
		},
	},
	{
		Name: "Receive error from UUIDs provider",
		BeforeTest: func(suite *uuidsWorkerTestSuite) {
			suite.Emitter.On("Emit", "skinsystem:before_request", mock.Anything)
			suite.Emitter.On("Emit", "skinsystem:after_request", mock.Anything, 500)
			err := errors.New("this is an error")
			suite.UuidsProvider.On("GetUuid", "mock_username").Return(nil, err)
		},
		AfterTest: func(suite *uuidsWorkerTestSuite, response *http.Response) {
			suite.Equal(500, response.StatusCode)
			suite.Equal("application/json", response.Header.Get("Content-Type"))
			body, _ := ioutil.ReadAll(response.Body)
			suite.JSONEq(`{
				"provider": "this is an error"
			}`, string(body))
		},
	},
	{
		Name: "Receive Too Many Requests from UUIDs provider",
		BeforeTest: func(suite *uuidsWorkerTestSuite) {
			suite.Emitter.On("Emit", "skinsystem:before_request", mock.Anything)
			suite.Emitter.On("Emit", "skinsystem:after_request", mock.Anything, 429)
			err := &mojang.TooManyRequestsError{}
			suite.UuidsProvider.On("GetUuid", "mock_username").Return(nil, err)
		},
		AfterTest: func(suite *uuidsWorkerTestSuite, response *http.Response) {
			suite.Equal(429, response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			suite.Empty(body)
		},
	},
}

func (suite *uuidsWorkerTestSuite) TestGetUUID() {
	for _, testCase := range getUuidTestsCases {
		suite.RunSubTest(testCase.Name, func() {
			testCase.BeforeTest(suite)

			req := httptest.NewRequest("GET", "http://chrly/api/worker/mojang-uuid/mock_username", nil)
			w := httptest.NewRecorder()

			suite.App.CreateHandler().ServeHTTP(w, req)

			testCase.AfterTest(suite, w.Result())
		})
	}
}
