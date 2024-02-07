package mojang

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var awaitDelay = 20 * time.Millisecond

type mojangUsernamesToUuidsRequestMock struct {
	mock.Mock
}

func (o *mojangUsernamesToUuidsRequestMock) UsernamesToUuids(usernames []string) ([]*ProfileInfo, error) {
	args := o.Called(usernames)
	var result []*ProfileInfo
	if casted, ok := args.Get(0).([]*ProfileInfo); ok {
		result = casted
	}

	return result, args.Error(1)
}

type batchUuidsProviderGetUuidResult struct {
	Result *ProfileInfo
	Error  error
}

type batchUuidsProviderTestSuite struct {
	suite.Suite

	Provider *BatchUuidsProvider

	MojangApi *mojangUsernamesToUuidsRequestMock
}

func (s *batchUuidsProviderTestSuite) SetupTest() {
	s.MojangApi = &mojangUsernamesToUuidsRequestMock{}
	s.Provider = NewBatchUuidsProvider(
		s.MojangApi.UsernamesToUuids,
		3,
		awaitDelay,
		false,
	)
}

func (s *batchUuidsProviderTestSuite) TearDownTest() {
	s.MojangApi.AssertExpectations(s.T())
	s.Provider.StopQueue()
}

func (s *batchUuidsProviderTestSuite) GetUuidAsync(username string) <-chan *batchUuidsProviderGetUuidResult {
	return s.GetUuidAsyncWithCtx(context.Background(), username)
}

func (s *batchUuidsProviderTestSuite) GetUuidAsyncWithCtx(ctx context.Context, username string) <-chan *batchUuidsProviderGetUuidResult {
	startedChan := make(chan any)
	c := make(chan *batchUuidsProviderGetUuidResult, 1)
	go func() {
		close(startedChan)
		profile, err := s.Provider.GetUuid(ctx, username)
		c <- &batchUuidsProviderGetUuidResult{
			Result: profile,
			Error:  err,
		}
		close(c)
	}()

	<-startedChan

	return c
}

func (s *batchUuidsProviderTestSuite) TestGetUuidForFewUsernamesSuccessfully() {
	expectedUsernames := []string{"username1", "username2"}
	expectedResult1 := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username1"}
	expectedResult2 := &ProfileInfo{Id: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "username2"}

	s.MojangApi.On("UsernamesToUuids", expectedUsernames).Once().Return([]*ProfileInfo{
		expectedResult1,
		expectedResult2,
	}, nil)

	chan1 := s.GetUuidAsync("username1")
	chan2 := s.GetUuidAsync("username2")

	s.Require().Empty(chan1)
	s.Require().Empty(chan2)

	time.Sleep(time.Duration(float64(awaitDelay) * 1.5))

	result1 := <-chan1
	result2 := <-chan2

	s.Require().NoError(result1.Error)
	s.Require().Equal(expectedResult1, result1.Result)

	s.Require().NoError(result2.Error)
	s.Require().Equal(expectedResult2, result2.Result)

	// Await a few more iterations to ensure, that no requests will be performed when there are no additional tasks
	time.Sleep(awaitDelay * 3)
}

func (s *batchUuidsProviderTestSuite) TestGetUuidForManyUsernamesSplitByMultipleIterations() {
	var emptyResponse []string

	s.MojangApi.On("UsernamesToUuids", []string{"username1", "username2", "username3"}).Once().Return(emptyResponse, nil)
	s.MojangApi.On("UsernamesToUuids", []string{"username4"}).Once().Return(emptyResponse, nil)

	resultChan1 := s.GetUuidAsync("username1")
	resultChan2 := s.GetUuidAsync("username2")
	resultChan3 := s.GetUuidAsync("username3")
	resultChan4 := s.GetUuidAsync("username4")

	time.Sleep(time.Duration(float64(awaitDelay) * 1.5))

	s.Require().NotEmpty(resultChan1)
	s.Require().NotEmpty(resultChan2)
	s.Require().NotEmpty(resultChan3)
	s.Require().Empty(resultChan4)

	time.Sleep(time.Duration(float64(awaitDelay) * 1.5))

	s.Require().NotEmpty(resultChan4)
}

func (s *batchUuidsProviderTestSuite) TestGetUuidForManyUsernamesWhenOneOfContextIsDeadlined() {
	var emptyResponse []string

	s.MojangApi.On("UsernamesToUuids", []string{"username1", "username2", "username4"}).Once().Return(emptyResponse, nil)

	ctx, cancelCtx := context.WithCancel(context.Background())

	resultChan1 := s.GetUuidAsync("username1")
	resultChan2 := s.GetUuidAsync("username2")
	resultChan3 := s.GetUuidAsyncWithCtx(ctx, "username3")
	resultChan4 := s.GetUuidAsync("username4")

	cancelCtx()

	time.Sleep(time.Duration(float64(awaitDelay) * 0.5))

	s.Empty(resultChan1)
	s.Empty(resultChan2)
	s.NotEmpty(resultChan3, "canceled context must immediately release the job")
	s.Empty(resultChan4)

	result3 := <-resultChan3
	s.Nil(result3.Result)
	s.ErrorIs(result3.Error, context.Canceled)

	time.Sleep(awaitDelay)

	s.Require().NotEmpty(resultChan1)
	s.Require().NotEmpty(resultChan2)
	s.Require().NotEmpty(resultChan4)
}

func (s *batchUuidsProviderTestSuite) TestGetUuidForManyUsernamesFireOnFull() {
	s.Provider.fireOnFull = true

	var emptyResponse []string

	s.MojangApi.On("UsernamesToUuids", []string{"username1", "username2", "username3"}).Once().Return(emptyResponse, nil)
	s.MojangApi.On("UsernamesToUuids", []string{"username4"}).Once().Return(emptyResponse, nil)

	resultChan1 := s.GetUuidAsync("username1")
	resultChan2 := s.GetUuidAsync("username2")
	resultChan3 := s.GetUuidAsync("username3")
	resultChan4 := s.GetUuidAsync("username4")

	time.Sleep(time.Duration(float64(awaitDelay) * 0.5))

	s.Require().NotEmpty(resultChan1)
	s.Require().NotEmpty(resultChan2)
	s.Require().NotEmpty(resultChan3)
	s.Require().Empty(resultChan4)

	time.Sleep(time.Duration(float64(awaitDelay) * 1.5))

	s.Require().NotEmpty(resultChan4)
}

func (s *batchUuidsProviderTestSuite) TestGetUuidForFewUsernamesWithAnError() {
	expectedUsernames := []string{"username1", "username2"}
	expectedError := errors.New("mock error")

	s.MojangApi.On("UsernamesToUuids", expectedUsernames).Once().Return(nil, expectedError)

	resultChan1 := s.GetUuidAsync("username1")
	resultChan2 := s.GetUuidAsync("username2")

	result1 := <-resultChan1
	s.Assert().Nil(result1.Result)
	s.Assert().Equal(expectedError, result1.Error)

	result2 := <-resultChan2
	s.Assert().Nil(result2.Result)
	s.Assert().Equal(expectedError, result2.Error)
}

func TestBatchUuidsProvider(t *testing.T) {
	suite.Run(t, new(batchUuidsProviderTestSuite))
}
