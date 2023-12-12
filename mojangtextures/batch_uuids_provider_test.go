package mojangtextures

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

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

type manualStrategy struct {
	ch   chan *JobsIteration
	once sync.Once
	lock sync.Mutex
	jobs []*job
}

func (m *manualStrategy) Queue(job *job) {
	m.lock.Lock()
	m.jobs = append(m.jobs, job)
	m.lock.Unlock()
}

func (m *manualStrategy) GetJobs(_ context.Context) <-chan *JobsIteration {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.ch = make(chan *JobsIteration)

	return m.ch
}

func (m *manualStrategy) Iterate(countJobsToReturn int, countLeftJobsInQueue int) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.ch <- &JobsIteration{
		Jobs:  m.jobs[0:countJobsToReturn],
		Queue: countLeftJobsInQueue,
	}
}

type batchUuidsProviderGetUuidResult struct {
	Result *mojang.ProfileInfo
	Error  error
}

type batchUuidsProviderTestSuite struct {
	suite.Suite

	Provider *BatchUuidsProvider

	Emitter   *mockEmitter
	Strategy  *manualStrategy
	MojangApi *mojangUsernamesToUuidsRequestMock

	stop context.CancelFunc
}

func (suite *batchUuidsProviderTestSuite) GetUuidAsync(username string) <-chan *batchUuidsProviderGetUuidResult {
	s := make(chan struct{})
	// This dirty hack ensures, that the username will be queued before we return control to the caller.
	// It's needed to keep expected calls order and prevent cases when iteration happens before
	// all usernames will be queued.
	suite.Emitter.On("Emit",
		"mojang_textures:batch_uuids_provider:queued",
		username,
	).Once().Run(func(args mock.Arguments) {
		close(s)
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

func (suite *batchUuidsProviderTestSuite) SetupTest() {
	suite.Emitter = &mockEmitter{}
	suite.Strategy = &manualStrategy{}
	ctx, stop := context.WithCancel(context.Background())
	suite.stop = stop
	suite.MojangApi = &mojangUsernamesToUuidsRequestMock{}
	usernamesToUuids = suite.MojangApi.UsernamesToUuids

	suite.Provider = NewBatchUuidsProvider(ctx, suite.Strategy, suite.Emitter)
}

func (suite *batchUuidsProviderTestSuite) TearDownTest() {
	suite.stop()
	suite.Emitter.AssertExpectations(suite.T())
	suite.MojangApi.AssertExpectations(suite.T())
}

func TestBatchUuidsProvider(t *testing.T) {
	suite.Run(t, new(batchUuidsProviderTestSuite))
}

func (suite *batchUuidsProviderTestSuite) TestGetUuidForFewUsernames() {
	expectedUsernames := []string{"username1", "username2"}
	expectedResult1 := &mojang.ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username1"}
	expectedResult2 := &mojang.ProfileInfo{Id: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "username2"}
	expectedResponse := []*mojang.ProfileInfo{expectedResult1, expectedResult2}

	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:round", expectedUsernames, 0).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:result", expectedUsernames, expectedResponse, nil).Once()

	suite.MojangApi.On("UsernamesToUuids", expectedUsernames).Once().Return([]*mojang.ProfileInfo{
		expectedResult1,
		expectedResult2,
	}, nil)

	resultChan1 := suite.GetUuidAsync("username1")
	resultChan2 := suite.GetUuidAsync("username2")

	suite.Strategy.Iterate(2, 0)

	result1 := <-resultChan1
	suite.Assert().Equal(expectedResult1, result1.Result)
	suite.Assert().Nil(result1.Error)

	result2 := <-resultChan2
	suite.Assert().Equal(expectedResult2, result2.Result)
	suite.Assert().Nil(result2.Error)
}

func (suite *batchUuidsProviderTestSuite) TestShouldNotSendRequestWhenNoJobsAreReturned() {
	//noinspection GoPreferNilSlice
	emptyUsernames := []string{}
	done := make(chan struct{})
	suite.Emitter.On("Emit",
		"mojang_textures:batch_uuids_provider:round",
		emptyUsernames,
		1,
	).Once().Run(func(args mock.Arguments) {
		close(done)
	})

	suite.GetUuidAsync("username") // Schedule one username to run the queue

	suite.Strategy.Iterate(0, 1) // Return no jobs and indicate that there is one job in queue

	<-done
}

// Test written for multiple usernames to ensure that the error
// will be returned for each iteration group
func (suite *batchUuidsProviderTestSuite) TestGetUuidForFewUsernamesWithAnError() {
	expectedUsernames := []string{"username1", "username2"}
	expectedError := &mojang.TooManyRequestsError{}
	var nilProfilesResponse []*mojang.ProfileInfo

	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:round", expectedUsernames, 0).Once()
	suite.Emitter.On("Emit", "mojang_textures:batch_uuids_provider:result", expectedUsernames, nilProfilesResponse, expectedError).Once()

	suite.MojangApi.On("UsernamesToUuids", expectedUsernames).Once().Return(nil, expectedError)

	resultChan1 := suite.GetUuidAsync("username1")
	resultChan2 := suite.GetUuidAsync("username2")

	suite.Strategy.Iterate(2, 0)

	result1 := <-resultChan1
	suite.Assert().Nil(result1.Result)
	suite.Assert().Equal(expectedError, result1.Error)

	result2 := <-resultChan2
	suite.Assert().Nil(result2.Result)
	suite.Assert().Equal(expectedError, result2.Error)
}

func TestPeriodicStrategy(t *testing.T) {
	t.Run("should return first job only after duration", func(t *testing.T) {
		d := 20 * time.Millisecond
		strategy := NewPeriodicStrategy(d, 10)
		j := &job{}
		strategy.Queue(j)

		ctx, cancel := context.WithCancel(context.Background())
		startedAt := time.Now()
		ch := strategy.GetJobs(ctx)
		iteration := <-ch
		durationBeforeResult := time.Now().Sub(startedAt)
		require.True(t, durationBeforeResult >= d)
		require.True(t, durationBeforeResult < d*2)

		require.Equal(t, []*job{j}, iteration.Jobs)
		require.Equal(t, 0, iteration.Queue)

		cancel()
	})

	t.Run("should return the configured batch size", func(t *testing.T) {
		strategy := NewPeriodicStrategy(0, 10)
		jobs := make([]*job, 15)
		for i := 0; i < 15; i++ {
			jobs[i] = &job{Username: strconv.Itoa(i)}
			strategy.Queue(jobs[i])
		}

		ctx, cancel := context.WithCancel(context.Background())
		ch := strategy.GetJobs(ctx)
		iteration := <-ch
		require.Len(t, iteration.Jobs, 10)
		require.Equal(t, jobs[0:10], iteration.Jobs)
		require.Equal(t, 5, iteration.Queue)

		cancel()
	})

	t.Run("should not return the next iteration until the previous one is finished", func(t *testing.T) {
		strategy := NewPeriodicStrategy(0, 10)
		strategy.Queue(&job{})

		ctx, cancel := context.WithCancel(context.Background())
		ch := strategy.GetJobs(ctx)
		iteration := <-ch
		require.Len(t, iteration.Jobs, 1)
		require.Equal(t, 0, iteration.Queue)

		time.Sleep(time.Millisecond) // Let strategy's internal loop to work (if the implementation is broken)

		select {
		case <-ch:
			require.Fail(t, "the previous iteration isn't marked as done")
		default:
			// ok
		}

		iteration.Done()

		time.Sleep(time.Millisecond) // Let strategy's internal loop to work

		select {
		case iteration = <-ch:
			// ok
		default:
			require.Fail(t, "iteration should be provided")
		}

		require.Empty(t, iteration.Jobs)
		require.Equal(t, 0, iteration.Queue)
		iteration.Done()

		cancel()
	})

	t.Run("each iteration should be returned only after the configured duration", func(t *testing.T) {
		d := 5 * time.Millisecond
		strategy := NewPeriodicStrategy(d, 10)
		ctx, cancel := context.WithCancel(context.Background())
		ch := strategy.GetJobs(ctx)
		for i := 0; i < 3; i++ {
			startedAt := time.Now()
			iteration := <-ch
			durationBeforeResult := time.Now().Sub(startedAt)
			require.True(t, durationBeforeResult >= d)
			require.True(t, durationBeforeResult < d*2)

			require.Empty(t, iteration.Jobs)
			require.Equal(t, 0, iteration.Queue)

			// Sleep for at least doubled duration before calling Done() to check,
			// that this duration isn't included into the next iteration time
			time.Sleep(d * 2)
			iteration.Done()
		}

		cancel()
	})
}

func TestFullBusStrategy(t *testing.T) {
	t.Run("should provide iteration immediately when the batch size exceeded", func(t *testing.T) {
		jobs := make([]*job, 10)
		for i := 0; i < 10; i++ {
			jobs[i] = &job{}
		}

		d := 20 * time.Millisecond
		strategy := NewFullBusStrategy(d, 10)
		ctx, cancel := context.WithCancel(context.Background())
		ch := strategy.GetJobs(ctx)

		done := make(chan struct{})
		go func() {
			defer close(done)
			select {
			case iteration := <-ch:
				require.Len(t, iteration.Jobs, 10)
				require.Equal(t, 0, iteration.Queue)
			case <-time.After(d):
				require.Fail(t, "iteration should be provided immediately")
			}
		}()

		for _, j := range jobs {
			strategy.Queue(j)
		}

		<-done

		cancel()
	})

	t.Run("should provide iteration after duration if batch size isn't exceeded", func(t *testing.T) {
		jobs := make([]*job, 9)
		for i := 0; i < 9; i++ {
			jobs[i] = &job{}
		}

		d := 20 * time.Millisecond
		strategy := NewFullBusStrategy(d, 10)
		ctx, cancel := context.WithCancel(context.Background())

		startedAt := time.Now()
		ch := strategy.GetJobs(ctx)

		done := make(chan struct{})
		go func() {
			defer close(done)
			iteration := <-ch
			duration := time.Now().Sub(startedAt)
			require.True(t, duration >= d, fmt.Sprintf("has %d, expected %d", duration, d))
			require.True(t, duration < d*2)
			require.Equal(t, jobs, iteration.Jobs)
			require.Equal(t, 0, iteration.Queue)
		}()

		for _, j := range jobs {
			strategy.Queue(j)
		}

		<-done

		cancel()
	})

	t.Run("should provide iteration as soon as the bus is full, without waiting for the previous iteration to finish", func(t *testing.T) {
		d := 20 * time.Millisecond
		strategy := NewFullBusStrategy(d, 10)
		ctx, cancel := context.WithCancel(context.Background())
		ch := strategy.GetJobs(ctx)

		done := make(chan struct{})
		go func() {
			defer close(done)
			for i := 0; i < 3; i++ {
				time.Sleep(5 * time.Millisecond) // See comment below
				select {
				case iteration := <-ch:
					require.Len(t, iteration.Jobs, 10)
					// Don't assert iteration.Queue length since it might be unstable
					// Don't call iteration.Done()
				case <-time.After(d):
					t.Errorf("iteration should be provided as soon as the bus is full")
					return
				}
			}

			// Scheduled 31 tasks. 3 iterations should be performed immediately
			// and should be executed only after timeout. The timeout above is used
			// to increase overall time to ensure, that timer resets on every iteration

			startedAt := time.Now()
			iteration := <-ch
			duration := time.Now().Sub(startedAt)
			require.True(t, duration >= d)
			require.True(t, duration < d*2)
			require.Len(t, iteration.Jobs, 1)
			require.Equal(t, 0, iteration.Queue)
		}()

		for i := 0; i < 31; i++ {
			strategy.Queue(&job{})
		}

		<-done

		cancel()
	})
}
