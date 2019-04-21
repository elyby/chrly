package queue

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/elyby/chrly/api/mojang"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net"
	"strings"
	"syscall"
	"testing"
	"time"
)

type mojangApiMocks struct {
	mock.Mock
}

func (o *mojangApiMocks) UsernamesToUuids(usernames []string) ([]*mojang.ProfileInfo, error) {
	args := o.Called(usernames)
	var result []*mojang.ProfileInfo
	if casted, ok := args.Get(0).([]*mojang.ProfileInfo); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (o *mojangApiMocks) UuidToTextures(uuid string, signed bool) (*mojang.SignedTexturesResponse, error) {
	args := o.Called(uuid, signed)
	var result *mojang.SignedTexturesResponse
	if casted, ok := args.Get(0).(*mojang.SignedTexturesResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) GetUuid(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *mockStorage) StoreUuid(username string, uuid string) {
	m.Called(username, uuid)
}

func (m *mockStorage) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	args := m.Called(uuid)
	var result *mojang.SignedTexturesResponse
	if casted, ok := args.Get(0).(*mojang.SignedTexturesResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *mockStorage) StoreTextures(textures *mojang.SignedTexturesResponse) {
	m.Called(textures)
}

type queueTestSuite struct {
	suite.Suite
	Queue     *JobsQueue
	Storage   *mockStorage
	MojangApi *mojangApiMocks
	Iterate   func()

	iterateChan chan bool
	done        func()
}

func (suite *queueTestSuite) SetupSuite() {
	uuidsQueuePeriod = 0
}

func (suite *queueTestSuite) SetupTest() {
	suite.Storage = &mockStorage{}

	suite.Queue = &JobsQueue{Storage: suite.Storage}

	suite.iterateChan = make(chan bool)
	forever = func() bool {
		return <-suite.iterateChan
	}

	suite.Iterate = func() {
		suite.iterateChan <- true
	}

	suite.done = func() {
		suite.iterateChan <- false
	}

	suite.MojangApi = new(mojangApiMocks)
	usernamesToUuids = suite.MojangApi.UsernamesToUuids
	uuidToTextures = suite.MojangApi.UuidToTextures
}

func (suite *queueTestSuite) TearDownTest() {
	suite.done()
	suite.MojangApi.AssertExpectations(suite.T())
	suite.Storage.AssertExpectations(suite.T())
}

func (suite *queueTestSuite) TestReceiveTexturesForOneUsernameWithoutAnyCache() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}

	suite.Storage.On("GetUuid", "maksimkurb").Once().Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "maksimkurb", "0d252b7218b648bfb86c2ae476954d32").Once()
	suite.Storage.On("GetTextures", "0d252b7218b648bfb86c2ae476954d32").Once().Return(nil, &ValueNotFound{})
	suite.Storage.On("StoreTextures", expectedResult).Once()
	suite.MojangApi.On("UsernamesToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(expectedResult, nil)

	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	result := <-resultChan
	suite.Assert().Equal(expectedResult, result)
}

func (suite *queueTestSuite) TestReceiveTexturesForFewUsernamesWithoutAnyCache() {
	expectedResult1 := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}
	expectedResult2 := &mojang.SignedTexturesResponse{Id: "4566e69fc90748ee8d71d7ba5aa00d20", Name: "Thinkofdeath"}

	suite.Storage.On("GetUuid", "maksimkurb").Once().Return("", &ValueNotFound{})
	suite.Storage.On("GetUuid", "Thinkofdeath").Once().Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "maksimkurb", "0d252b7218b648bfb86c2ae476954d32").Once()
	suite.Storage.On("StoreUuid", "Thinkofdeath", "4566e69fc90748ee8d71d7ba5aa00d20").Once()
	suite.Storage.On("GetTextures", "0d252b7218b648bfb86c2ae476954d32").Once().Return(nil, &ValueNotFound{})
	suite.Storage.On("GetTextures", "4566e69fc90748ee8d71d7ba5aa00d20").Once().Return(nil, &ValueNotFound{})
	suite.Storage.On("StoreTextures", expectedResult1).Once()
	suite.Storage.On("StoreTextures", expectedResult2).Once()
	suite.MojangApi.On("UsernamesToUuids", []string{"maksimkurb", "Thinkofdeath"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
		{Id: "4566e69fc90748ee8d71d7ba5aa00d20", Name: "Thinkofdeath"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(expectedResult1, nil)
	suite.MojangApi.On("UuidToTextures", "4566e69fc90748ee8d71d7ba5aa00d20", true).Once().Return(expectedResult2, nil)

	resultChan1 := suite.Queue.GetTexturesForUsername("maksimkurb")
	resultChan2 := suite.Queue.GetTexturesForUsername("Thinkofdeath")

	suite.Iterate()

	suite.Assert().Equal(expectedResult1, <-resultChan1)
	suite.Assert().Equal(expectedResult2, <-resultChan2)
}

func (suite *queueTestSuite) TestReceiveTexturesForUsernameWithCachedUuid() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}

	suite.Storage.On("GetUuid", "maksimkurb").Once().Return("0d252b7218b648bfb86c2ae476954d32", nil)
	// Storage.StoreUuid shouldn't be called
	suite.Storage.On("GetTextures", "0d252b7218b648bfb86c2ae476954d32").Once().Return(nil, &ValueNotFound{})
	suite.Storage.On("StoreTextures", expectedResult).Once()
	// MojangApi.UsernamesToUuids shouldn't be called
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(expectedResult, nil)

	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	// Note that there is no iteration

	result := <-resultChan
	suite.Assert().Equal(expectedResult, result)
}

func (suite *queueTestSuite) TestReceiveTexturesForUsernameWithFullyCachedResult() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}

	suite.Storage.On("GetUuid", "maksimkurb").Once().Return("0d252b7218b648bfb86c2ae476954d32", nil)
	// Storage.StoreUuid shouldn't be called
	suite.Storage.On("GetTextures", "0d252b7218b648bfb86c2ae476954d32").Once().Return(expectedResult, nil)
	// Storage.StoreTextures shouldn't be called
	// MojangApi.UsernamesToUuids shouldn't be called
	// MojangApi.UuidToTextures shouldn't be called

	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	// Note that there is no iteration

	result := <-resultChan
	suite.Assert().Equal(expectedResult, result)
}

func (suite *queueTestSuite) TestReceiveTexturesForUsernameWithCachedUnknownUuid() {
	suite.Storage.On("GetUuid", "maksimkurb").Once().Return("", nil)
	// Storage.StoreUuid shouldn't be called
	// Storage.GetTextures shouldn't be called
	// Storage.StoreTextures shouldn't be called
	// MojangApi.UsernamesToUuids shouldn't be called
	// MojangApi.UuidToTextures shouldn't be called

	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	// Note that there is no iteration

	suite.Assert().Nil(<-resultChan)
}

func (suite *queueTestSuite) TestReceiveTexturesForMoreThan100Usernames() {
	usernames := make([]string, 120, 120)
	for i := 0; i < 120; i++ {
		usernames[i] = randStr(8)
	}

	suite.Storage.On("GetUuid", mock.Anything).Times(120).Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", mock.Anything, "").Times(120) // if username is not compared to uuid, then receive ""
	// Storage.GetTextures and Storage.SetTextures shouldn't be called
	suite.MojangApi.On("UsernamesToUuids", usernames[0:100]).Once().Return([]*mojang.ProfileInfo{}, nil)
	suite.MojangApi.On("UsernamesToUuids", usernames[100:120]).Once().Return([]*mojang.ProfileInfo{}, nil)

	for _, username := range usernames {
		suite.Queue.GetTexturesForUsername(username)
	}

	suite.Iterate()
	suite.Iterate()
}

func (suite *queueTestSuite) TestReceiveTexturesForTheSameUsernames() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}

	suite.Storage.On("GetUuid", "maksimkurb").Twice().Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "maksimkurb", "0d252b7218b648bfb86c2ae476954d32").Once()
	suite.Storage.On("GetTextures", "0d252b7218b648bfb86c2ae476954d32").Once().Return(nil, &ValueNotFound{})
	suite.Storage.On("StoreTextures", expectedResult).Once()
	suite.MojangApi.On("UsernamesToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(expectedResult, nil)

	resultChan1 := suite.Queue.GetTexturesForUsername("maksimkurb")
	resultChan2 := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	suite.Assert().Equal(expectedResult, <-resultChan1)
	suite.Assert().Equal(expectedResult, <-resultChan2)
}

func (suite *queueTestSuite) TestReceiveTexturesForUsernameThatAlreadyProcessing() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}

	suite.Storage.On("GetUuid", "maksimkurb").Twice().Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "maksimkurb", "0d252b7218b648bfb86c2ae476954d32").Once()
	suite.Storage.On("GetTextures", "0d252b7218b648bfb86c2ae476954d32").Once().Return(nil, &ValueNotFound{})
	suite.Storage.On("StoreTextures", expectedResult).Once()
	suite.MojangApi.On("UsernamesToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).
		Once().
		After(10*time.Millisecond). // Simulate long round trip
		Return(expectedResult, nil)

	resultChan1 := suite.Queue.GetTexturesForUsername("maksimkurb")

	// Note that for entire test there is only one iteration
	suite.Iterate()

	// Let it meet delayed UuidToTextures request
	time.Sleep(5 * time.Millisecond)

	resultChan2 := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Assert().Equal(expectedResult, <-resultChan1)
	suite.Assert().Equal(expectedResult, <-resultChan2)
}

func (suite *queueTestSuite) TestDoNothingWhenNoTasks() {
	suite.Storage.On("GetUuid", "maksimkurb").Once().Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "maksimkurb", "").Once()
	// Storage.GetTextures and Storage.StoreTextures shouldn't be called
	suite.MojangApi.On("UsernamesToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{}, nil)

	// Perform first iteration and await it finish
	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	suite.Assert().Nil(<-resultChan)

	// Let it to perform a few more iterations to ensure, that there is no calls to external APIs
	suite.Iterate()
	suite.Iterate()
}

type timeoutError struct {
}

func (*timeoutError) Error() string   { return "timeout error" }
func (*timeoutError) Timeout() bool   { return true }
func (*timeoutError) Temporary() bool { return false }

var expectedErrors = []error{
	&mojang.BadRequestError{},
	&mojang.TooManyRequestsError{},
	&mojang.ServerError{},
	&timeoutError{},
	&net.OpError{Op: "read"},
	&net.OpError{Op: "dial"},
	syscall.ECONNREFUSED,
}

func (suite *queueTestSuite) TestShouldNotPanicWhenExpectedErrorReturnedFromUsernameToUuidRequest() {
	for _, err := range expectedErrors {
		suite.Storage.On("GetUuid", "maksimkurb").Once().Return("", &ValueNotFound{})
		suite.MojangApi.On("UsernamesToUuids", []string{"maksimkurb"}).Once().Return(nil, err)
		resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")
		suite.Iterate()
		suite.Assert().Nil(<-resultChan)
		suite.MojangApi.AssertExpectations(suite.T())
		suite.MojangApi.ExpectedCalls = nil // https://github.com/stretchr/testify/issues/558#issuecomment-372112364
	}
}

func (suite *queueTestSuite) TestShouldNotPanicWhenExpectedErrorReturnedFromUuidToTexturesRequest() {
	for _, err := range expectedErrors {
		suite.Storage.On("GetUuid", "maksimkurb").Once().Return("", &ValueNotFound{})
		suite.Storage.On("StoreUuid", "maksimkurb", "0d252b7218b648bfb86c2ae476954d32").Once()
		suite.Storage.On("GetTextures", "0d252b7218b648bfb86c2ae476954d32").Once().Return(nil, &ValueNotFound{})
		// Storage.StoreTextures shouldn't be called
		suite.MojangApi.On("UsernamesToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
			{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
		}, nil)
		suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(nil, err)
		resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")
		suite.Iterate()
		suite.Assert().Nil(<-resultChan)
		suite.MojangApi.AssertExpectations(suite.T())
		suite.MojangApi.ExpectedCalls = nil // https://github.com/stretchr/testify/issues/558#issuecomment-372112364
	}
}

func (suite *queueTestSuite) TestReceiveTexturesForNotAllowedMojangUsername() {
	resultChan := suite.Queue.GetTexturesForUsername("Not allowed")
	suite.Assert().Nil(<-resultChan)
}

func TestJobsQueueSuite(t *testing.T) {
	suite.Run(t, new(queueTestSuite))
}

var replacer = strings.NewReplacer("-", "_", "=", "")

// https://stackoverflow.com/a/50581165
func randStr(len int) string {
	buff := make([]byte, len)
	_, _ = rand.Read(buff)
	str := replacer.Replace(base64.URLEncoding.EncodeToString(buff))

	// Base 64 can be longer than len
	return str[:len]
}
