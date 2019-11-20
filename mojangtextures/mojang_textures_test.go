package mojangtextures

import (
	"errors"
	"net"
	"net/url"
	"sync"
	"syscall"
	"testing"
	"time"

	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/api/mojang"
	mocks "github.com/elyby/chrly/tests"
)

func TestBroadcaster(t *testing.T) {
	t.Run("GetOrAppend", func(t *testing.T) {
		t.Run("first call when username didn't exist before should return true", func(t *testing.T) {
			assert := testify.New(t)

			broadcaster := createBroadcaster()
			channel := make(chan *broadcastResult)
			isFirstListener := broadcaster.AddListener("mock", channel)

			assert.True(isFirstListener)
			listeners, ok := broadcaster.listeners["mock"]
			assert.True(ok)
			assert.Len(listeners, 1)
			assert.Equal(channel, listeners[0])
		})

		t.Run("subsequent calls should return false", func(t *testing.T) {
			assert := testify.New(t)

			broadcaster := createBroadcaster()
			channel1 := make(chan *broadcastResult)
			isFirstListener := broadcaster.AddListener("mock", channel1)

			assert.True(isFirstListener)

			channel2 := make(chan *broadcastResult)
			isFirstListener = broadcaster.AddListener("mock", channel2)

			assert.False(isFirstListener)

			channel3 := make(chan *broadcastResult)
			isFirstListener = broadcaster.AddListener("mock", channel3)

			assert.False(isFirstListener)
		})
	})

	t.Run("BroadcastAndRemove", func(t *testing.T) {
		t.Run("should broadcast to all listeners and remove the key", func(t *testing.T) {
			assert := testify.New(t)

			broadcaster := createBroadcaster()
			channel1 := make(chan *broadcastResult)
			channel2 := make(chan *broadcastResult)
			broadcaster.AddListener("mock", channel1)
			broadcaster.AddListener("mock", channel2)

			result := &broadcastResult{}
			broadcaster.BroadcastAndRemove("mock", result)

			assert.Equal(result, <-channel1)
			assert.Equal(result, <-channel2)

			channel3 := make(chan *broadcastResult)
			isFirstListener := broadcaster.AddListener("mock", channel3)
			assert.True(isFirstListener)
		})

		t.Run("call on not exists username", func(t *testing.T) {
			assert := testify.New(t)

			assert.NotPanics(func() {
				broadcaster := createBroadcaster()
				broadcaster.BroadcastAndRemove("mock", &broadcastResult{})
			})
		})
	})
}

type mockUuidsProvider struct {
	mock.Mock
}

func (m *mockUuidsProvider) GetUuid(username string) (*mojang.ProfileInfo, error) {
	args := m.Called(username)
	var result *mojang.ProfileInfo
	if casted, ok := args.Get(0).(*mojang.ProfileInfo); ok {
		result = casted
	}

	return result, args.Error(1)
}

type mockTexturesProvider struct {
	mock.Mock
}

func (m *mockTexturesProvider) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	args := m.Called(uuid)
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

func (m *mockStorage) StoreUuid(username string, uuid string) error {
	args := m.Called(username, uuid)
	return args.Error(0)
}

func (m *mockStorage) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	args := m.Called(uuid)
	var result *mojang.SignedTexturesResponse
	if casted, ok := args.Get(0).(*mojang.SignedTexturesResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *mockStorage) StoreTextures(uuid string, textures *mojang.SignedTexturesResponse) {
	m.Called(uuid, textures)
}

type providerTestSuite struct {
	suite.Suite
	Provider         *Provider
	UuidsProvider    *mockUuidsProvider
	TexturesProvider *mockTexturesProvider
	Storage          *mockStorage
	Logger           *mocks.WdMock
}

func (suite *providerTestSuite) SetupTest() {
	suite.UuidsProvider = &mockUuidsProvider{}
	suite.TexturesProvider = &mockTexturesProvider{}
	suite.Storage = &mockStorage{}
	suite.Logger = &mocks.WdMock{}

	suite.Provider = &Provider{
		UuidsProvider:    suite.UuidsProvider,
		TexturesProvider: suite.TexturesProvider,
		Storage:          suite.Storage,
		Logger:           suite.Logger,
	}
}

func (suite *providerTestSuite) TearDownTest() {
	// time.Sleep(10 * time.Millisecond) // Add delay to let finish all goroutines before assert mocks calls
	suite.UuidsProvider.AssertExpectations(suite.T())
	suite.TexturesProvider.AssertExpectations(suite.T())
	suite.Storage.AssertExpectations(suite.T())
	suite.Logger.AssertExpectations(suite.T())
}

func TestProvider(t *testing.T) {
	suite.Run(t, new(providerTestSuite))
}

func (suite *providerTestSuite) TestGetForUsernameWithoutAnyCache() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.Logger.On("IncCounter", "mojang_textures.request", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.uuid_hit", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.textures_hit", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.result_time", mock.Anything).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil)
	suite.Storage.On("StoreTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult).Once()

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(&mojang.ProfileInfo{
		Id:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "username",
	}, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(err)
	suite.Assert().Equal(expectedResult, result)
}

func (suite *providerTestSuite) TestGetForUsernameWithCachedUuid() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.Logger.On("IncCounter", "mojang_textures.request", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.cache_hit", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.textures_hit", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.result_time", mock.Anything).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil)
	suite.Storage.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, &ValueNotFound{})
	suite.Storage.On("StoreTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult).Once()

	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(err)
	suite.Assert().Equal(expectedResult, result)
}

func (suite *providerTestSuite) TestGetForUsernameWithFullyCachedResult() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.Logger.On("IncCounter", "mojang_textures.request", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.cache_hit", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.textures.cache_hit", int64(1)).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil)
	suite.Storage.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(err)
	suite.Assert().Equal(expectedResult, result)
}

func (suite *providerTestSuite) TestGetForUsernameWithCachedUnknownUuid() {
	suite.Logger.On("IncCounter", "mojang_textures.request", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.cache_hit_nil", int64(1)).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(result)
	suite.Assert().Nil(err)
}

func (suite *providerTestSuite) TestGetForUsernameWhichHasNoMojangAccount() {
	suite.Logger.On("IncCounter", "mojang_textures.request", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.uuid_miss", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.result_time", mock.Anything).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "username", "").Once().Return(nil)

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(nil, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(err)
	suite.Assert().Nil(result)
}

func (suite *providerTestSuite) TestGetForUsernameWhichHasMojangAccountButHasNoMojangSkin() {
	var expectedResult *mojang.SignedTexturesResponse

	suite.Logger.On("IncCounter", "mojang_textures.request", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.uuid_hit", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.textures_miss", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.result_time", mock.Anything).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil)
	suite.Storage.On("StoreTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult).Once()

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(&mojang.ProfileInfo{
		Id:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "username",
	}, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Equal(expectedResult, result)
	suite.Assert().Nil(err)
}

func (suite *providerTestSuite) TestGetForTheSameUsernames() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.Logger.On("IncCounter", "mojang_textures.request", int64(1)).Twice()
	suite.Logger.On("IncCounter", "mojang_textures.already_scheduled", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.uuid_hit", int64(1)).Once()
	suite.Logger.On("IncCounter", "mojang_textures.usernames.textures_hit", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.result_time", mock.Anything).Once()

	suite.Storage.On("GetUuid", "username").Twice().Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil)
	suite.Storage.On("StoreTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult).Once()

	// If possible, than remove this .After call
	suite.UuidsProvider.On("GetUuid", "username").Once().After(time.Millisecond).Return(&mojang.ProfileInfo{
		Id:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "username",
	}, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	results := make([]*mojang.SignedTexturesResponse, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			textures, _ := suite.Provider.GetForUsername("username")
			results[i] = textures
			wg.Done()
		}(i)
	}
	wg.Wait()

	suite.Assert().Equal(expectedResult, results[0])
	suite.Assert().Equal(expectedResult, results[1])
}

func (suite *providerTestSuite) TestGetForNotAllowedMojangUsername() {
	suite.Logger.On("IncCounter", "mojang_textures.invalid_username", int64(1)).Once()

	result, err := suite.Provider.GetForUsername("Not allowed")
	suite.Assert().Error(err, "invalid username")
	suite.Assert().Nil(result)
}

type timeoutError struct {
}

func (*timeoutError) Error() string   { return "timeout error" }
func (*timeoutError) Timeout() bool   { return true }
func (*timeoutError) Temporary() bool { return false }

var expectedErrors = []error{
	&mojang.BadRequestError{},
	&mojang.ForbiddenError{},
	&mojang.TooManyRequestsError{},
	&mojang.ServerError{},
	&timeoutError{},
	&url.Error{Op: "GET", URL: "http://localhost"},
	&net.OpError{Op: "read"},
	&net.OpError{Op: "dial"},
	syscall.ECONNREFUSED,
}

func (suite *providerTestSuite) TestShouldNotLogErrorWhenExpectedErrorReturnedFromUsernameToUuidRequest() {
	suite.Logger.On("IncCounter", mock.Anything, mock.Anything)
	suite.Logger.On("RecordTimer", mock.Anything, mock.Anything)
	suite.Logger.On("Debug", ":name: Got response error :err", mock.Anything, mock.Anything).Times(len(expectedErrors))
	suite.Logger.On("Warning", ":name: Got 400 Bad Request :err", mock.Anything, mock.Anything).Once()
	suite.Logger.On("Warning", ":name: Got 403 Forbidden :err", mock.Anything, mock.Anything).Once()
	suite.Logger.On("Warning", ":name: Got 429 Too Many Requests :err", mock.Anything, mock.Anything).Once()

	suite.Storage.On("GetUuid", "username").Return("", &ValueNotFound{})

	for _, err := range expectedErrors {
		suite.UuidsProvider.On("GetUuid", "username").Once().Return(nil, err)

		result, err := suite.Provider.GetForUsername("username")
		suite.Assert().Nil(result)
		suite.Assert().NotNil(err)
		suite.UuidsProvider.AssertExpectations(suite.T())
		suite.UuidsProvider.ExpectedCalls = nil // https://github.com/stretchr/testify/issues/558#issuecomment-372112364
	}
}

func (suite *providerTestSuite) TestShouldLogEmergencyOnUnexpectedErrorReturnedFromUsernameToUuidRequest() {
	suite.Logger.On("IncCounter", mock.Anything, mock.Anything)
	suite.Logger.On("RecordTimer", mock.Anything, mock.Anything)
	suite.Logger.On("Debug", ":name: Got response error :err", mock.Anything, mock.Anything).Once()
	suite.Logger.On("Emergency", ":name: Unknown Mojang response error: :err", mock.Anything, mock.Anything).Once()

	suite.Storage.On("GetUuid", "username").Return("", &ValueNotFound{})

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(nil, errors.New("unexpected error"))

	result, err := suite.Provider.GetForUsername("username")
	suite.Assert().Nil(result)
	suite.Assert().NotNil(err)
}

func (suite *providerTestSuite) TestShouldNotLogErrorWhenExpectedErrorReturnedFromUuidToTexturesRequest() {
	suite.Logger.On("IncCounter", mock.Anything, mock.Anything)
	suite.Logger.On("RecordTimer", mock.Anything, mock.Anything)
	suite.Logger.On("Debug", ":name: Got response error :err", mock.Anything, mock.Anything).Times(len(expectedErrors))
	suite.Logger.On("Warning", ":name: Got 400 Bad Request :err", mock.Anything, mock.Anything).Once()
	suite.Logger.On("Warning", ":name: Got 403 Forbidden :err", mock.Anything, mock.Anything).Once()
	suite.Logger.On("Warning", ":name: Got 429 Too Many Requests :err", mock.Anything, mock.Anything).Once()

	suite.Storage.On("GetUuid", "username").Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Return(nil)
	// suite.Storage.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Return(nil, &ValueNotFound{})
	// suite.Storage.On("StoreTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", (*mojang.SignedTexturesResponse)(nil))

	for _, err := range expectedErrors {
		suite.UuidsProvider.On("GetUuid", "username").Once().Return(&mojang.ProfileInfo{
			Id:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Name: "username",
		}, nil)
		suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, err)

		result, err := suite.Provider.GetForUsername("username")
		suite.Assert().Nil(result)
		suite.Assert().NotNil(err)
		suite.UuidsProvider.AssertExpectations(suite.T())
		suite.TexturesProvider.AssertExpectations(suite.T())
		suite.UuidsProvider.ExpectedCalls = nil    // https://github.com/stretchr/testify/issues/558#issuecomment-372112364
		suite.TexturesProvider.ExpectedCalls = nil // https://github.com/stretchr/testify/issues/558#issuecomment-372112364
	}
}

func (suite *providerTestSuite) TestShouldLogEmergencyOnUnexpectedErrorReturnedFromUuidToTexturesRequest() {
	suite.Logger.On("IncCounter", mock.Anything, mock.Anything)
	suite.Logger.On("RecordTimer", mock.Anything, mock.Anything)
	suite.Logger.On("Debug", ":name: Got response error :err", mock.Anything, mock.Anything).Once()
	suite.Logger.On("Emergency", ":name: Unknown Mojang response error: :err", mock.Anything, mock.Anything).Once()

	suite.Storage.On("GetUuid", "username").Return("", &ValueNotFound{})
	suite.Storage.On("StoreUuid", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Return(nil)

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(&mojang.ProfileInfo{
		Id:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "username",
	}, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, errors.New("unexpected error"))

	result, err := suite.Provider.GetForUsername("username")
	suite.Assert().Nil(result)
	suite.Assert().NotNil(err)
}
