package mojangtextures

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"sync"
	"testing"
	"time"

	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/api/mojang"
	mocks "github.com/elyby/chrly/tests"
)

func TestJobsQueue(t *testing.T) {
	createQueue := func() *jobsQueue {
		queue := &jobsQueue{}
		queue.New()

		return queue
	}

	t.Run("Enqueue", func(t *testing.T) {
		assert := testify.New(t)

		s := createQueue()
		s.Enqueue(&jobItem{username: "username1"})
		s.Enqueue(&jobItem{username: "username2"})
		s.Enqueue(&jobItem{username: "username3"})

		assert.Equal(3, s.Size())
	})

	t.Run("Dequeue", func(t *testing.T) {
		assert := testify.New(t)

		s := createQueue()
		s.Enqueue(&jobItem{username: "username1"})
		s.Enqueue(&jobItem{username: "username2"})
		s.Enqueue(&jobItem{username: "username3"})
		s.Enqueue(&jobItem{username: "username4"})

		items := s.Dequeue(2)
		assert.Len(items, 2)
		assert.Equal("username1", items[0].username)
		assert.Equal("username2", items[1].username)
		assert.Equal(2, s.Size())

		items = s.Dequeue(40)
		assert.Len(items, 2)
		assert.Equal("username3", items[0].username)
		assert.Equal("username4", items[1].username)
	})
}

// This is really stupid test just to get 100% coverage on this package :)
func TestBatchUuidsProvider_forever(t *testing.T) {
	testify.True(t, forever())
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

type batchUuidsProviderGetUuidResult struct {
	Result *mojang.ProfileInfo
	Error  error
}

type batchUuidsProviderTestSuite struct {
	suite.Suite

	Provider     *BatchUuidsProvider
	GetUuidAsync func(username string) chan *batchUuidsProviderGetUuidResult

	Logger    *mocks.WdMock
	MojangApi *mojangUsernamesToUuidsRequestMock

	Iterate     func()
	done        func()
	iterateChan chan bool
}

func (suite *batchUuidsProviderTestSuite) SetupTest() {
	suite.Logger = &mocks.WdMock{}

	suite.Provider = &BatchUuidsProvider{
		Logger:         suite.Logger,
		IterationDelay: 0,
		IterationSize:  10,
	}

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

	var lock sync.Mutex
	suite.GetUuidAsync = func(username string) chan *batchUuidsProviderGetUuidResult {
		lock.Lock()
		defer lock.Unlock()

		c := make(chan *batchUuidsProviderGetUuidResult)
		s := make(chan int)
		go func() {
			s <- 0
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
	time.Sleep(10 * time.Millisecond) // Add delay to let finish all goroutines before assert mocks calls
	suite.MojangApi.AssertExpectations(suite.T())
	suite.Logger.AssertExpectations(suite.T())
}

func TestBatchUuidsProvider(t *testing.T) {
	suite.Run(t, new(batchUuidsProviderTestSuite))
}

func (suite *batchUuidsProviderTestSuite) TestGetUuidForOneUsername() {
	expectedResult := &mojang.ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.Logger.On("IncCounter", "mojang_textures.usernames.queued", int64(1)).Once()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.iteration_size", int64(1)).Once()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.queue_size", int64(0)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.round_time", mock.Anything).Once()

	suite.MojangApi.On("UsernamesToUuids", []string{"username"}).Once().Return([]*mojang.ProfileInfo{expectedResult}, nil)

	resultChan := suite.GetUuidAsync("username")

	suite.Iterate()

	result := <-resultChan
	suite.Assert().Equal(expectedResult, result.Result)
	suite.Assert().Nil(result.Error)
}

func (suite *batchUuidsProviderTestSuite) TestGetUuidForTwoUsernames() {
	expectedResult1 := &mojang.ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username1"}
	expectedResult2 := &mojang.ProfileInfo{Id: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "username2"}

	suite.Logger.On("IncCounter", "mojang_textures.usernames.queued", int64(1)).Twice()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.iteration_size", int64(2)).Once()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.queue_size", int64(0)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.round_time", mock.Anything).Once()

	suite.MojangApi.On("UsernamesToUuids", []string{"username1", "username2"}).Once().Return([]*mojang.ProfileInfo{
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

	suite.Logger.On("IncCounter", "mojang_textures.usernames.queued", int64(1)).Times(12)
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.iteration_size", int64(10)).Once()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.iteration_size", int64(2)).Once()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.queue_size", int64(2)).Once()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.queue_size", int64(0)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.round_time", mock.Anything).Twice()

	suite.MojangApi.On("UsernamesToUuids", usernames[0:10]).Once().Return([]*mojang.ProfileInfo{}, nil)
	suite.MojangApi.On("UsernamesToUuids", usernames[10:12]).Once().Return([]*mojang.ProfileInfo{}, nil)

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
	suite.Logger.On("IncCounter", "mojang_textures.usernames.queued", int64(1)).Once()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.iteration_size", int64(1)).Once()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.iteration_size", int64(0)).Twice()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.queue_size", int64(0)).Times(3)
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.round_time", mock.Anything)

	suite.MojangApi.On("UsernamesToUuids", []string{"username"}).Once().Return([]*mojang.ProfileInfo{}, nil)

	// Perform first iteration and await it finish
	resultChan := suite.GetUuidAsync("username")

	suite.Iterate()

	result := <-resultChan
	suite.Assert().Nil(result.Result)
	suite.Assert().Nil(result.Error)

	// Let it to perform a few more iterations to ensure, that there is no calls to external APIs
	suite.Iterate()
	suite.Iterate()
}

func (suite *batchUuidsProviderTestSuite) TestGetUuidForTwoUsernamesWithAnError() {
	expectedError := &mojang.TooManyRequestsError{}

	suite.Logger.On("IncCounter", "mojang_textures.usernames.queued", int64(1)).Twice()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.iteration_size", int64(2)).Once()
	suite.Logger.On("UpdateGauge", "mojang_textures.usernames.queue_size", int64(0)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.round_time", mock.Anything).Once()

	suite.MojangApi.On("UsernamesToUuids", []string{"username1", "username2"}).Once().Return(nil, expectedError)

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
