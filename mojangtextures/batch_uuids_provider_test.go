package mojangtextures

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/api/mojang"
)

func TestJobsQueue(t *testing.T) {
	t.Run("Enqueue", func(t *testing.T) {
		s := newJobsQueue()
		require.Equal(t, 1, s.Enqueue(&job{Username: "username1"}))
		require.Equal(t, 2, s.Enqueue(&job{Username: "username2"}))
		require.Equal(t, 3, s.Enqueue(&job{Username: "username3"}))
	})

	t.Run("Dequeue", func(t *testing.T) {
		s := newJobsQueue()
		s.Enqueue(&job{Username: "username1"})
		s.Enqueue(&job{Username: "username2"})
		s.Enqueue(&job{Username: "username3"})
		s.Enqueue(&job{Username: "username4"})
		s.Enqueue(&job{Username: "username5"})

		items, queueLen := s.Dequeue(2)
		require.Len(t, items, 2)
		require.Equal(t, 3, queueLen)
		require.Equal(t, "username1", items[0].Username)
		require.Equal(t, "username2", items[1].Username)

		items, queueLen = s.Dequeue(40)
		require.Len(t, items, 3)
		require.Equal(t, 0, queueLen)
		require.Equal(t, "username3", items[0].Username)
		require.Equal(t, "username4", items[1].Username)
		require.Equal(t, "username5", items[2].Username)
	})
}

type mojangUsernamesToUuidsRequestMock struct {
	mock.Mock
}

func (o *mojangUsernamesToUuidsRequestMock) UsernamesToUuids(usernames []string) ([]*mojang.ProfileInfo, error) {
	args := o.Called(usernames)
	var result []*mojang.ProfileInfo
	if casted, ok := args.Get(0).([]*mojang.ProfileInfo); ok {
		result = casted
	}

	return result, args.Error(1)
}

type queueStrategyMock struct {
	mock.Mock
	ch chan *JobsIteration
}

func (m *queueStrategyMock) Queue(job *job) {
	m.Called(job)
}

func (m *queueStrategyMock) GetJobs(abort context.Context) <-chan *JobsIteration {
	m.Called(abort)
	return m.ch
}

func (m *queueStrategyMock) PushIteration(iteration *JobsIteration) {
	m.ch <- iteration
}

type batchUuidsProviderGetUuidResult struct {
	Result *mojang.ProfileInfo
	Error  error
}

type batchUuidsProviderTestSuite struct {
	suite.Suite

	Provider     *BatchUuidsProvider
	GetUuidAsync func(username string) chan *batchUuidsProviderGetUuidResult

	Emitter   *mockEmitter
	MojangApi *mojangUsernamesToUuidsRequestMock
}

func (suite *batchUuidsProviderTestSuite) SetupTest() {
	suite.Emitter = &mockEmitter{}

	suite.Provider = &BatchUuidsProvider{
		// Emitter:        suite.Emitter,
		// IterationDelay: 0,
		// IterationSize:  10,
	}

	suite.iterateChan = make(chan bool)
	// forever = func() bool {
	// 	return <-suite.iterateChan
	// }

	suite.Iterate = func() {
		suite.iterateChan <- true
	}

	suite.done = func() {
		suite.iterateChan <- false
	}

	suite.GetUuidAsync = func(username string) chan *batchUuidsProviderGetUuidResult {
		s := make(chan bool)
		// This dirty hack ensures, that the username will be queued before we return control to the caller.
		// It's needed to keep expected calls order and prevent cases when iteration happens before all usernames
		// will be queued.
		suite.Emitter.On("Emit",
			"mojang_textures:batch_uuids_provider:queued",
			username,
		).Once().Run(func(args mock.Arguments) {
			s <- true
		})

		c := make(chan *batchUuidsProviderGetUuidResult)
		go func() {
			profile, err := suite.Provider.GetUuid(username)
			c <- &batchUuidsProviderGetUuidResult{
				Result: profile,
				Error:  err,
			}
		}()

		<-s

		return c
	}

	suite.MojangApi = &mojangUsernamesToUuidsRequestMock{}
	usernamesToUuids = suite.MojangApi.UsernamesToUuids
}

func (suite *batchUuidsProviderTestSuite) TearDownTest() {
	suite.done()
	suite.Emitter.AssertExpectations(suite.T())
	suite.MojangApi.AssertExpectations(suite.T())
}

func TestBatchUuidsProvider(t *testing.T) {
	suite.Run(t, new(batchUuidsProviderTestSuite))
}

func (suite *batchUuidsProviderTestSuite) TestGetUuidForOneUsername() {
	expectedUsernames := []string{"username"}
	expectedResult := &mojang.ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	expectedResponse := []*mojang.ProfileInfo{expectedResult}

	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:before_round").Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:round", expectedUsernames, 0).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:result", expectedUsernames, expectedResponse, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:after_round").Once()

	suite.MojangApi.On("UsernamesToUuids", expectedUsernames).Once().Return([]*mojang.ProfileInfo{expectedResult}, nil)

	resultChan := suite.GetUuidAsync("username")

	suite.Iterate()

	result := <-resultChan
	suite.Assert().Equal(expectedResult, result.Result)
	suite.Assert().Nil(result.Error)
}

func (suite *batchUuidsProviderTestSuite) TestGetUuidForTwoUsernames() {
	expectedUsernames := []string{"username1", "username2"}
	expectedResult1 := &mojang.ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username1"}
	expectedResult2 := &mojang.ProfileInfo{Id: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "username2"}
	expectedResponse := []*mojang.ProfileInfo{expectedResult1, expectedResult2}

	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:before_round").Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:round", expectedUsernames, 0).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:result", expectedUsernames, expectedResponse, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:after_round").Once()

	suite.MojangApi.On("UsernamesToUuids", expectedUsernames).Once().Return([]*mojang.ProfileInfo{
		expectedResult1,
		expectedResult2,
	}, nil)

	resultChan1 := suite.GetUuidAsync("username1")
	resultChan2 := suite.GetUuidAsync("username2")

	suite.Iterate()

	result1 := <-resultChan1
	suite.Assert().Equal(expectedResult1, result1.Result)
	suite.Assert().Nil(result1.Error)

	result2 := <-resultChan2
	suite.Assert().Equal(expectedResult2, result2.Result)
	suite.Assert().Nil(result2.Error)
}

func (suite *batchUuidsProviderTestSuite) TestGetUuidForMoreThan10Usernames() {
	usernames := make([]string, 12)
	for i := 0; i < cap(usernames); i++ {
		usernames[i] = randStr(8)
	}

	// In this test we're not testing response, so always return an empty resultset
	expectedResponse := []*mojang.ProfileInfo{}

	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:before_round").Twice()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:round", usernames[0:10], 2).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:result", usernames[0:10], expectedResponse, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:round", usernames[10:12], 0).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:result", usernames[10:12], expectedResponse, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:after_round").Twice()

	suite.MojangApi.On("UsernamesToUuids", usernames[0:10]).Once().Return(expectedResponse, nil)
	suite.MojangApi.On("UsernamesToUuids", usernames[10:12]).Once().Return(expectedResponse, nil)

	channels := make([]chan *batchUuidsProviderGetUuidResult, len(usernames))
	for i, username := range usernames {
		channels[i] = suite.GetUuidAsync(username)
	}

	suite.Iterate()
	suite.Iterate()

	for _, channel := range channels {
		<-channel
	}
}

func (suite *batchUuidsProviderTestSuite) TestDoNothingWhenNoTasks() {
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:before_round").Times(3)
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:round", []string{"username"}, 0).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:result", []string{"username"}, mock.Anything, nil).Once()
	var nilStringSlice []string
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:round", nilStringSlice, 0).Twice()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:after_round").Times(3)

	suite.MojangApi.On("UsernamesToUuids", []string{"username"}).Once().Return([]*mojang.ProfileInfo{}, nil)

	// Perform first iteration and await it finishes
	resultChan := suite.GetUuidAsync("username")

	suite.Iterate()

	result := <-resultChan
	suite.Assert().Nil(result.Result)
	suite.Assert().Nil(result.Error)

	// Let it to perform a few more iterations to ensure, that there are no calls to external APIs
	suite.Iterate()
	suite.Iterate()
}

func (suite *batchUuidsProviderTestSuite) TestGetUuidForTwoUsernamesWithAnError() {
	expectedUsernames := []string{"username1", "username2"}
	expectedError := &mojang.TooManyRequestsError{}
	var nilProfilesResponse []*mojang.ProfileInfo

	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:before_round").Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:round", expectedUsernames, 0).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:result", expectedUsernames, nilProfilesResponse, expectedError).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:after_round").Once()

	suite.MojangApi.On("UsernamesToUuids", expectedUsernames).Once().Return(nil, expectedError)

	resultChan1 := suite.GetUuidAsync("username1")
	resultChan2 := suite.GetUuidAsync("username2")

	suite.Iterate()

	result1 := <-resultChan1
	suite.Assert().Nil(result1.Result)
	suite.Assert().Equal(expectedError, result1.Error)

	result2 := <-resultChan2
	suite.Assert().Nil(result2.Result)
	suite.Assert().Equal(expectedError, result2.Error)
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
