package mojang

import (
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

func (m *mockUuidsProvider) GetUuid(username string) (*ProfileInfo, error) {
	args := m.Called(username)
	var result *ProfileInfo
	if casted, ok := args.Get(0).(*ProfileInfo); ok {
		result = casted
	}

	return result, args.Error(1)
}

type TexturesProviderMock struct {
	mock.Mock
}

func (m *TexturesProviderMock) GetTextures(uuid string) (*ProfileResponse, error) {
	args := m.Called(uuid)
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

func (suite *providerTestSuite) SetupTest() {
	suite.UuidsProvider = &mockUuidsProvider{}
	suite.TexturesProvider = &TexturesProviderMock{}

	suite.Provider = &MojangTexturesProvider{
		UuidsProvider:    suite.UuidsProvider,
		TexturesProvider: suite.TexturesProvider,
	}
}

func (suite *providerTestSuite) TearDownTest() {
	suite.UuidsProvider.AssertExpectations(suite.T())
	suite.TexturesProvider.AssertExpectations(suite.T())
}

func (suite *providerTestSuite) TestGetForValidUsernameSuccessfully() {
	expectedProfile := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	expectedResult := &ProfileResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(expectedProfile, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedResult, result)
}

func (suite *providerTestSuite) TestGetForUsernameWhichHasNoMojangAccount() {
	suite.UuidsProvider.On("GetUuid", "username").Once().Return(nil, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().NoError(err)
	suite.Assert().Nil(result)
}

func (suite *providerTestSuite) TestGetForUsernameWhichHasMojangAccountButHasNoMojangSkin() {
	expectedProfile := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(expectedProfile, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, nil)

	result, err := suite.Provider.GetForUsername("username")

	suite.Assert().NoError(err)
	suite.Assert().Nil(result)
}

func (suite *providerTestSuite) TestGetForTheSameUsername() {
	expectedProfile := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	expectedResult := &ProfileResponse{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}

	awaitChan := make(chan time.Time)

	// If possible, then remove this .After call
	suite.UuidsProvider.On("GetUuid", "username").Once().WaitUntil(awaitChan).Return(expectedProfile, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(expectedResult, nil)

	results := make([]*ProfileResponse, 2)
	var wgStarted sync.WaitGroup
	var wgDone sync.WaitGroup
	for i := 0; i < 2; i++ {
		wgStarted.Add(1)
		wgDone.Add(1)
		go func(i int) {
			wgStarted.Done()
			textures, _ := suite.Provider.GetForUsername("username")
			results[i] = textures
			wgDone.Done()
		}(i)
	}

	wgStarted.Wait()
	close(awaitChan)
	wgDone.Wait()

	suite.Assert().Equal(expectedResult, results[0])
	suite.Assert().Equal(expectedResult, results[1])
}

func (suite *providerTestSuite) TestGetForNotAllowedMojangUsername() {
	result, err := suite.Provider.GetForUsername("Not allowed")
	suite.Assert().ErrorIs(err, InvalidUsername)
	suite.Assert().Nil(result)
}

func (suite *providerTestSuite) TestGetErrorFromUuidsProvider() {
	err := errors.New("mock error")
	suite.UuidsProvider.On("GetUuid", "username").Once().Return(nil, err)

	result, resErr := suite.Provider.GetForUsername("username")
	suite.Assert().Nil(result)
	suite.Assert().Equal(err, resErr)
}

func (suite *providerTestSuite) TestGetErrorFromTexturesProvider() {
	expectedProfile := &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "username"}
	err := errors.New("mock error")

	suite.UuidsProvider.On("GetUuid", "username").Once().Return(expectedProfile, nil)
	suite.TexturesProvider.On("GetTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil, err)

	result, resErr := suite.Provider.GetForUsername("username")
	suite.Assert().Nil(result)
	suite.Assert().Equal(err, resErr)
}

func TestProvider(t *testing.T) {
	suite.Run(t, new(providerTestSuite))
}

func TestNilProvider_GetForUsername(t *testing.T) {
	provider := &NilProvider{}
	result, err := provider.GetForUsername("username")
	require.Nil(t, result)
	require.NoError(t, err)
}
