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

type QueueTestSuite struct {
	suite.Suite
	Queue     *JobsQueue
	MojangApi *MojangApiMocks
	Iterate   func()

	iterateChan chan bool
	done        func()
}

func (suite *QueueTestSuite) SetupSuite() {
	delay = 0
}

func (suite *QueueTestSuite) SetupTest() {
	suite.Queue = &JobsQueue{}

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
}

func (suite *QueueTestSuite) TestReceiveTexturesForOneUsername() {
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(
		&mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
		nil,
	)

	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	result := <-resultChan
	if suite.Assert().NotNil(result) {
		suite.Assert().Equal("0d252b7218b648bfb86c2ae476954d32", result.Id)
		suite.Assert().Equal("maksimkurb", result.Name)
	}
}

func (suite *QueueTestSuite) TestReceiveTexturesForFewUsernames() {
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb", "Thinkofdeath"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
		{Id: "4566e69fc90748ee8d71d7ba5aa00d20", Name: "Thinkofdeath"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(
		&mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
		nil,
	)
	suite.MojangApi.On("UuidToTextures", "4566e69fc90748ee8d71d7ba5aa00d20", true).Once().Return(
		&mojang.SignedTexturesResponse{Id: "4566e69fc90748ee8d71d7ba5aa00d20", Name: "Thinkofdeath"},
		nil,
	)

	resultChan1 := suite.Queue.GetTexturesForUsername("maksimkurb")
	resultChan2 := suite.Queue.GetTexturesForUsername("Thinkofdeath")

	suite.Iterate()

	suite.Assert().NotNil(<-resultChan1)
	suite.Assert().NotNil(<-resultChan2)
}

func (suite *QueueTestSuite) TestReceiveTexturesForMoreThan100Usernames() {
	usernames := make([]string, 120, 120)
	for i := 0; i < 120; i++ {
		usernames[i] = randStr(8)
	}

	suite.MojangApi.On("UsernameToUuids", usernames[0:100]).Once().Return([]*mojang.ProfileInfo{}, nil)
	suite.MojangApi.On("UsernameToUuids", usernames[100:120]).Once().Return([]*mojang.ProfileInfo{}, nil)

	for _, username := range usernames {
		suite.Queue.GetTexturesForUsername(username)
	}

	suite.Iterate()
	suite.Iterate()
}

func (suite *QueueTestSuite) TestReceiveTexturesForTheSameUsernames() {
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).Once().Return(
		&mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
		nil,
	)

	resultChan1 := suite.Queue.GetTexturesForUsername("maksimkurb")
	resultChan2 := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	result1 := <-resultChan1
	result2 := <-resultChan2

	if suite.Assert().NotNil(result1) {
		suite.Assert().Equal("0d252b7218b648bfb86c2ae476954d32", result1.Id)
		suite.Assert().Equal("maksimkurb", result1.Name)

		suite.Assert().Equal(result1, result2)
	}
}

func (suite *QueueTestSuite) TestReceiveTexturesForUsernameThatAlreadyProcessing() {
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return([]*mojang.ProfileInfo{
		{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
	}, nil)
	suite.MojangApi.On("UuidToTextures", "0d252b7218b648bfb86c2ae476954d32", true).
		Once().
		After(10*time.Millisecond). // Simulate long round trip
		Return(
			&mojang.SignedTexturesResponse{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
			nil,
		)

	resultChan1 := suite.Queue.GetTexturesForUsername("maksimkurb")

	// Note that for entire test there is only one iteration
	suite.Iterate()

	// Let it meet delayed UuidToTextures request
	time.Sleep(5 * time.Millisecond)

	resultChan2 := suite.Queue.GetTexturesForUsername("maksimkurb")

	result1 := <-resultChan1
	result2 := <-resultChan2

	if suite.Assert().NotNil(result1) {
		suite.Assert().Equal("0d252b7218b648bfb86c2ae476954d32", result1.Id)
		suite.Assert().Equal("maksimkurb", result1.Name)

		suite.Assert().Equal(result1, result2)
	}
}

func (suite *QueueTestSuite) TestDoNothingWhenNoTasks() {
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
	suite.MojangApi.On("UsernameToUuids", []string{"maksimkurb"}).Once().Return(nil, &mojang.TooManyRequestsError{})

	resultChan := suite.Queue.GetTexturesForUsername("maksimkurb")

	suite.Iterate()

	suite.Assert().Nil(<-resultChan)
}

func (suite *QueueTestSuite) TestHandle429ResponseWhenRequestingUsersTextures() {
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
