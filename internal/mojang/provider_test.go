package mojang

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type mockUuidsProvider struct {
	mock.Mock
}

func (m *mockUuidsProvider) GetUuid(ctx context.Context, username string) (*ProfileInfo, error) {
	args := m.Called(ctx, username)
	var result *ProfileInfo
	if casted, ok := args.Get(0).(*ProfileInfo); ok {
		result = casted
	}

	return result, args.Error(1)
}

type TexturesProviderMock struct {
	mock.Mock
}

func (m *TexturesProviderMock) GetTextures(ctx context.Context, uuid string) (*ProfileResponse, error) {
	args := m.Called(ctx, uuid)
	var result *ProfileResponse
	if casted, ok := args.Get(0).(*ProfileResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

type providerTestSuite struct {
	suite.Suite
	Provider         *MojangTexturesProvider
	UuidsProvider    *mockUuidsProvider
	TexturesProvider *TexturesProviderMock
}

func (s *providerTestSuite) SetupTest() {
	s.UuidsProvider = &mockUuidsProvider{}
	s.TexturesProvider = &TexturesProviderMock{}

	s.Provider = &MojangTexturesProvider{
		UuidsProvider:    s.UuidsProvider,
		TexturesProvider: s.TexturesProvider,
	}
}

func (s *providerTestSuite) TearDownTest() {
	s.UuidsProvider.AssertExpectations(s.T())
	s.TexturesProvider.AssertExpectations(s.T())
}

func (s *providerTestSuite) TestGetForValidUsernameSuccessfully() {
	expectedProfile := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	expectedResult := &ProfileResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	ctx := context.Background()

	s.UuidsProvider.On("GetUuid", ctx, "username").Once().Return(expectedProfile, nil)
	s.TexturesProvider.On("GetTextures", ctx, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	result, err := s.Provider.GetForUsername(ctx, "username")

	s.NoError(err)
	s.Same(expectedResult, result)
}

func (s *providerTestSuite) TestGetForUsernameWhichHasNoMojangAccount() {
	s.UuidsProvider.On("GetUuid", mock.Anything, "username").Once().Return(nil, nil)

	result, err := s.Provider.GetForUsername(context.Background(), "username")

	s.NoError(err)
	s.Nil(result)
}

func (s *providerTestSuite) TestGetForUsernameWhichHasMojangAccountButHasNoMojangSkin() {
	expectedProfile := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	s.UuidsProvider.On("GetUuid", mock.Anything, "username").Once().Return(expectedProfile, nil)
	s.TexturesProvider.On("GetTextures", mock.Anything, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, nil)

	result, err := s.Provider.GetForUsername(context.Background(), "username")

	s.NoError(err)
	s.Nil(result)
}

func (s *providerTestSuite) TestGetForTheSameUsernameInRow() {
	expectedProfile := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	expectedResult := &ProfileResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	awaitChan := make(chan time.Time)

	s.UuidsProvider.On("GetUuid", mock.Anything, "username").Once().WaitUntil(awaitChan).Return(expectedProfile, nil)
	s.TexturesProvider.On("GetTextures", mock.Anything, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	results := make([]*ProfileResponse, 2)
	var wgStarted sync.WaitGroup
	var wgDone sync.WaitGroup
	for i := 0; i < 2; i++ {
		wgStarted.Add(1)
		wgDone.Add(1)
		go func(i int) {
			wgStarted.Done()
			textures, _ := s.Provider.GetForUsername(context.Background(), "username")
			results[i] = textures
			wgDone.Done()
		}(i)
	}

	wgStarted.Wait()
	close(awaitChan)
	wgDone.Wait()

	s.Same(expectedResult, results[0])
	s.Same(expectedResult, results[1])
}

func (s *providerTestSuite) TestGetForTheSameUsernameOneAfterAnother() {
	expectedProfile := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	expectedResult := &ProfileResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	s.UuidsProvider.On("GetUuid", mock.Anything, "username").Times(2).Return(expectedProfile, nil)
	s.TexturesProvider.On("GetTextures", mock.Anything, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Times(2).Return(expectedResult, nil)

	// Just ensure that providers will be called twice
	_, _ = s.Provider.GetForUsername(context.Background(), "username")
	time.Sleep(time.Millisecond * 20)
	_, _ = s.Provider.GetForUsername(context.Background(), "username")
}

func (s *providerTestSuite) TestGetForNotAllowedMojangUsername() {
	result, err := s.Provider.GetForUsername(context.Background(), "Not allowed")
	s.ErrorIs(err, InvalidUsername)
	s.Nil(result)
}

func (s *providerTestSuite) TestGetErrorFromUuidsProvider() {
	err := errors.New("mock error")
	s.UuidsProvider.On("GetUuid", mock.Anything, "username").Once().Return(nil, err)

	result, resErr := s.Provider.GetForUsername(context.Background(), "username")
	s.Nil(result)
	s.Equal(err, resErr)
}

func (s *providerTestSuite) TestGetErrorFromTexturesProvider() {
	expectedProfile := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	err := errors.New("mock error")

	s.UuidsProvider.On("GetUuid", mock.Anything, "username").Once().Return(expectedProfile, nil)
	s.TexturesProvider.On("GetTextures", mock.Anything, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, err)

	result, resErr := s.Provider.GetForUsername(context.Background(), "username")
	s.Nil(result)
	s.Same(err, resErr)
}

func TestProvider(t *testing.T) {
	suite.Run(t, new(providerTestSuite))
}

func TestNilProvider_GetForUsername(t *testing.T) {
	provider := &NilProvider{}
	result, err := provider.GetForUsername(context.Background(), "username")
	require.Nil(t, result)
	require.NoError(t, err)
}
