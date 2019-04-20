package queue

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/elyby/chrly/api/mojang"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MojangApiMocks struct {
	mock.Mock
}

func (o *MojangApiMocks) UsernameToUuids(usernames []string) ([]*mojang.ProfileInfo, error) {
	args := o.Called(usernames)
	var result []*mojang.ProfileInfo
	if casted, ok := args.Get(0).([]*mojang.ProfileInfo); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (o *MojangApiMocks) UuidToTextures(uuid string, signed bool) (*mojang.SignedTexturesResponse, error) {
	args := o.Called(uuid, signed)
	var result *mojang.SignedTexturesResponse
	if casted, ok := args.Get(0).(*mojang.SignedTexturesResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Get(username string) *mojang.SignedTexturesResponse {
	args := m.Called(username)
	var result *mojang.SignedTexturesResponse
	if casted, ok := args.Get(0).(*mojang.SignedTexturesResponse); ok {
		result = casted
	}

	return result
}

func (m *MockStorage) Set(textures *mojang.SignedTexturesResponse) {
	m.Called(textures)
}

type QueueTestSuite struct {
	suite.Suite
	Queue     *JobsQueue
	Storage   *MockStorage
	MojangApi *MojangApiMocks
	Iterate   func()

	iterateChan chan bool
	done        func()
}

func (suite *QueueTestSuite) SetupSuite() {
	delay = 0
}

func (suite *QueueTestSuite) SetupTest() {
	suite.Storage = &MockStorage{}

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

	suite.MojangApi = new(MojangApiMocks)
	usernamesToUuids = suite.MojangApi.UsernameToUuids
	uuidToTextures = suite.MojangApi.UuidToTextures
}

func (suite *QueueTestSuite) TearDownTest() {
	suite.done()
	suite.MojangApi.AssertExpectations(suite.T())
	suite.Storage.AssertExpectations(suite.T())
}

func (suite *QueueTestSuite) TestReceiveTexturesForOneUsername() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}

	suite.Storage.On("Get", mock.Anything).Return(nil)
	suite.Storage.On("Set", expectedResult).Once()
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(expectedResult, nil)

	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	result := <-resultChan
	suite.Assert().Equal(expectedResult, result)
}

func (suite *QueueTestSuite) TestReceiveTexturesForFewUsernames() {
	expectedResult1 := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}
	expectedResult2 := &mojang.SignedTexturesResponse{Id: "4566e69fc90748ee8d71d7ba5aa00d20", Name: "Thinkofdeath"}

	suite.Storage.On("Get", mock.Anything).Return(nil)
	suite.Storage.On("Set", expectedResult1).Once()
	suite.Storage.On("Set", expectedResult2).Once()
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb", "Thinkofdeath"}).Once().Return([]*mojang.ProfileInfo{
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

func (suite *QueueTestSuite) TestReceiveTexturesForMoreThan100Usernames() {
	usernames := make([]string, 120, 120)
	for i := 0; i < 120; i++ {
		usernames[i] = randStr(8)
	}

	suite.Storage.On("Get", mock.Anything).Times(120).Return(nil)
	// Storage.Set shouldn't be called
	suite.MojangApi.On("UsernameToUuids", usernames[0:100]).Once().Return([]*mojang.ProfileInfo{}, nil)
	suite.MojangApi.On("UsernameToUuids", usernames[100:120]).Once().Return([]*mojang.ProfileInfo{}, nil)

	for _, username := range usernames {
		suite.Queue.GetTexturesForUsername(username)
	}

	suite.Iterate()
	suite.Iterate()
}

func (suite *QueueTestSuite) TestReceiveTexturesForTheSameUsernames() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}

	suite.Storage.On("Get", mock.Anything).Twice().Return(nil)
	suite.Storage.On("Set", expectedResult).Once()
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(expectedResult, nil)

	resultChan1 := suite.Queue.GetTexturesForUsername("maksimkurb")
	resultChan2 := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	suite.Assert().Equal(expectedResult, <-resultChan1)
	suite.Assert().Equal(expectedResult, <-resultChan2)
}

func (suite *QueueTestSuite) TestReceiveTexturesForUsernameThatAlreadyProcessing() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"}

	suite.Storage.On("Get", mock.Anything).Return(nil)
	suite.Storage.On("Set", expectedResult).Once()
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
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

func (suite *QueueTestSuite) TestDoNothingWhenNoTasks() {
	suite.Storage.On("Get", mock.Anything).Return(nil)
	// Storage.Set shouldn't be called
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{}, nil)

	// Perform first iteration and await it finish
	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	suite.Assert().Nil(<-resultChan)

	// Let it to perform a few more iterations to ensure, that there is no calls to external APIs
	suite.Iterate()
	suite.Iterate()
}

func (suite *QueueTestSuite) TestHandle429ResponseWhenExchangingUsernamesToUuids() {
	suite.Storage.On("Get", mock.Anything).Return(nil)
	// Storage.Set shouldn't be called
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return(nil, &mojang.TooManyRequestsError{})

	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	suite.Assert().Nil(<-resultChan)
}

func (suite *QueueTestSuite) TestHandle429ResponseWhenRequestingUsersTextures() {
	suite.Storage.On("Get", mock.Anything).Return(nil)
	// Storage.Set shouldn't be called
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(
		nil,
		&mojang.TooManyRequestsError{},
	)

	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	suite.Assert().Nil(<-resultChan)
}

func TestJobsQueueSuite(t *testing.T) {
	suite.Run(t, new(QueueTestSuite))
}

// https://stackoverflow.com/a/50581165
func randStr(len int) string {
	buff := make([]byte, len)
	_, _ = rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)

	// Base 64 can be longer than len
	return str[:len]
}
