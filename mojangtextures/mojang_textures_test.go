package mojangtextures

import (
	"errors"
	"sync"
	"testing"
	"time"

	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/api/mojang"
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

type mockEmitter struct {
	mock.Mock
}

func (e *mockEmitter) Emit(name string, args ...interface{}) {
	e.Called(append([]interface{}{name}, args...)...)
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

func (m *mockStorage) GetUuid(username string) (string, bool, error) {
	args := m.Called(username)
	return args.String(0), args.Bool(1), args.Error(2)
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
	Emitter          *mockEmitter
	UuidsProvider    *mockUuidsProvider
	TexturesProvider *mockTexturesProvider
	Storage          *mockStorage
}

func (suite *providerTestSuite) SetupTest() {
	suite.Emitter = &mockEmitter{}
	suite.UuidsProvider = &mockUuidsProvider{}
	suite.TexturesProvider = &mockTexturesProvider{}
	suite.Storage = &mockStorage{}

	suite.Provider = &Provider{
		Emitter:          suite.Emitter,
		UUIDsProvider:    suite.UuidsProvider,
		TexturesProvider: suite.TexturesProvider,
		Storage:          suite.Storage,
	}
}

func (suite *providerTestSuite) TearDownTest() {
	suite.Emitter.AssertExpectations(suite.T())
	suite.UuidsProvider.AssertExpectations(suite.T())
	suite.TexturesProvider.AssertExpectations(suite.T())
	suite.Storage.AssertExpectations(suite.T())
}

func TestProvider(t *testing.T) {
	suite.Run(t, new(providerTestSuite))
}

func (suite *providerTestSuite) TestGetForUsernameWithoutAnyCache() {
	expectedProfile := &mojang.ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	expectedResult := &mojang.SignedTexturesResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "", false, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:before_result", "username", "").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_call", "username", expectedProfile, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:after_result", "username", expectedResult, nil).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", false, nil)
	suite.Storage.On("StoreUuid", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil)
	suite.Storage.On("StoreTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult).Once()

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(expectedProfile, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(err)
	suite.Assert().Equal(expectedResult, result)
}

func (suite *providerTestSuite) TestGetForUsernameWithCachedUuid() {
	var expectedCachedTextures *mojang.SignedTexturesResponse
	expectedResult := &mojang.SignedTexturesResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_cache", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_cache", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedCachedTextures, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:before_result", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:after_result", "username", expectedResult, nil).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, nil)
	suite.Storage.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, nil)
	suite.Storage.On("StoreTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult).Once()

	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(err)
	suite.Assert().Equal(expectedResult, result)
}

func (suite *providerTestSuite) TestGetForUsernameWithFullyCachedResult() {
	expectedResult := &mojang.SignedTexturesResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_cache", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_cache", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult, nil).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, nil)
	suite.Storage.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(err)
	suite.Assert().Equal(expectedResult, result)
}

func (suite *providerTestSuite) TestGetForUsernameWithCachedUnknownUuid() {
	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "", true, nil).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", true, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(result)
	suite.Assert().Nil(err)
}

func (suite *providerTestSuite) TestGetForUsernameWhichHasNoMojangAccount() {
	var expectedProfile *mojang.ProfileInfo
	var expectedResult *mojang.SignedTexturesResponse

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "", false, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:before_result", "username", "").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_call", "username", expectedProfile, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:after_result", "username", expectedResult, nil).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", false, nil)
	suite.Storage.On("StoreUuid", "username", "").Once().Return(nil)

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(nil, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(err)
	suite.Assert().Nil(result)
}

func (suite *providerTestSuite) TestGetForUsernameWhichHasMojangAccountButHasNoMojangSkin() {
	expectedProfile := &mojang.ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	var expectedResult *mojang.SignedTexturesResponse

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "", false, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:before_result", "username", "").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_call", "username", expectedProfile, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:after_result", "username", expectedResult, nil).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", false, nil)
	suite.Storage.On("StoreUuid", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil)
	suite.Storage.On("StoreTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult).Once()

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(expectedProfile, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Equal(expectedResult, result)
	suite.Assert().Nil(err)
}

// https://github.com/elyby/chrly/issues/29
func (suite *providerTestSuite) TestGetForUsernameWithCachedUuidThatHasBeenDisappeared() {
	expectedErr := &mojang.EmptyResponse{}
	expectedProfile := &mojang.ProfileInfo{Id: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "username"}
	var nilTexturesResponse *mojang.SignedTexturesResponse
	expectedResult := &mojang.SignedTexturesResponse{Id: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "username"}

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_cache", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_cache", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nilTexturesResponse, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:before_result", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nilTexturesResponse, expectedErr).Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_call", "username", expectedProfile, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_call", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_call", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", expectedResult, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:after_result", "username", expectedResult, nil).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, nil)
	suite.Storage.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, nil)
	suite.Storage.On("StoreUuid", "username", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb").Once().Return(nil)
	suite.Storage.On("StoreTextures", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", expectedResult).Once()

	suite.UuidsProvider.On("GetUuid", "username").Return(expectedProfile, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Return(nil, expectedErr)
	suite.TexturesProvider.On("GetTextures", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb").Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(err)
	suite.Assert().Equal(expectedResult, result)
}

func (suite *providerTestSuite) TestGetForTheSameUsernames() {
	expectedProfile := &mojang.ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	expectedResult := &mojang.SignedTexturesResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Twice()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Twice()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "", false, nil).Twice()
	suite.Emitter.On("Emit", "mojang_textures:already_processing", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:before_result", "username", "").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_call", "username", expectedProfile, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:after_result", "username", expectedResult, nil).Once()

	suite.Storage.On("GetUuid", "username").Twice().Return("", false, nil)
	suite.Storage.On("StoreUuid", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil)
	suite.Storage.On("StoreTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult).Once()

	// If possible, than remove this .After call
	suite.UuidsProvider.On("GetUuid", "username").Once().After(time.Millisecond).Return(expectedProfile, nil)
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
	result, err := suite.Provider.GetForUsername("Not allowed")
	suite.Assert().Error(err, "invalid username")
	suite.Assert().Nil(result)
}

func (suite *providerTestSuite) TestGetErrorFromUUIDsStorage() {
	expectedErr := errors.New("mock error")

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "", false, expectedErr).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", false, expectedErr)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().Nil(result)
	suite.Assert().Equal(expectedErr, err)
}

func (suite *providerTestSuite) TestGetErrorFromUuidsProvider() {
	var expectedProfile *mojang.ProfileInfo
	var expectedResult *mojang.SignedTexturesResponse
	err := errors.New("mock error")

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "", false, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:before_result", "username", "").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_call", "username", expectedProfile, err).Once()
	suite.Emitter.On("Emit", "mojang_textures:after_result", "username", expectedResult, err).Once()

	suite.Storage.On("GetUuid", "username").Once().Return("", false, nil)
	suite.UuidsProvider.On("GetUuid", "username").Once().Return(nil, err)

	result, resErr := suite.Provider.GetForUsername("username")
	suite.Assert().Nil(result)
	suite.Assert().Equal(err, resErr)
}

func (suite *providerTestSuite) TestGetErrorFromTexturesProvider() {
	expectedProfile := &mojang.ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	var expectedResult *mojang.SignedTexturesResponse
	err := errors.New("mock error")

	suite.Emitter.On("Emit", "mojang_textures:call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_cache", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_cache", "username", "", false, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:before_result", "username", "").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:before_call", "username").Once()
	suite.Emitter.On("Emit", "mojang_textures:usernames:after_call", "username", expectedProfile, nil).Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:before_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once()
	suite.Emitter.On("Emit", "mojang_textures:textures:after_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expectedResult, err).Once()
	suite.Emitter.On("Emit", "mojang_textures:after_result", "username", expectedResult, err).Once()

	suite.Storage.On("GetUuid", "username").Return("", false, nil)
	suite.Storage.On("StoreUuid", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Return(nil)
	suite.UuidsProvider.On("GetUuid", "username").Once().Return(expectedProfile, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, err)

	result, resErr := suite.Provider.GetForUsername("username")
	suite.Assert().Nil(result)
	suite.Assert().Equal(err, resErr)
}
