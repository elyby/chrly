package mojang

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var mockProfile = &ProfileInfo{Id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "UserName"}

type UuidsProviderMock struct {
	mock.Mock
}

func (m *UuidsProviderMock) GetUuid(username string) (*ProfileInfo, error) {
	args := m.Called(username)
	var result *ProfileInfo
	if casted, ok := args.Get(0).(*ProfileInfo); ok {
		result = casted
	}

	return result, args.Error(1)
}

type MojangUuidsStorageMock struct {
	mock.Mock
}

func (m *MojangUuidsStorageMock) GetUuidForMojangUsername(username string) (string, string, error) {
	args := m.Called(username)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MojangUuidsStorageMock) StoreMojangUuid(username string, uuid string) error {
	m.Called(username, uuid)
	return nil
}

type UuidsProviderWithCacheSuite struct {
	suite.Suite

	Original *UuidsProviderMock
	Storage  *MojangUuidsStorageMock
	Provider *UuidsProviderWithCache
}

func (s *UuidsProviderWithCacheSuite) SetupTest() {
	s.Original = &UuidsProviderMock{}
	s.Storage = &MojangUuidsStorageMock{}
	s.Provider = &UuidsProviderWithCache{
		Provider: s.Original,
		Storage:  s.Storage,
	}
}

func (s *UuidsProviderWithCacheSuite) TearDownTest() {
	s.Original.AssertExpectations(s.T())
	s.Storage.AssertExpectations(s.T())
}

func (s *UuidsProviderWithCacheSuite) TestUncachedSuccessfully() {
	s.Storage.On("GetUuidForMojangUsername", "username").Return("", "", nil)
	s.Storage.On("StoreMojangUuid", "UserName", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Once().Return(nil)

	s.Original.On("GetUuid", "username").Once().Return(mockProfile, nil)

	result, err := s.Provider.GetUuid("username")

	s.Require().NoError(err)
	s.Require().Equal(mockProfile, result)
}

func (s *UuidsProviderWithCacheSuite) TestUncachedNotExistsMojangUsername() {
	s.Storage.On("GetUuidForMojangUsername", "username").Return("", "", nil)
	s.Storage.On("StoreMojangUuid", "username", "").Once().Return(nil)

	s.Original.On("GetUuid", "username").Once().Return(nil, nil)

	result, err := s.Provider.GetUuid("username")

	s.Require().NoError(err)
	s.Require().Nil(result)
}

func (s *UuidsProviderWithCacheSuite) TestKnownCachedUsername() {
	s.Storage.On("GetUuidForMojangUsername", "username").Return("mock-uuid", "UserName", nil)

	result, err := s.Provider.GetUuid("username")

	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Require().Equal("UserName", result.Name)
	s.Require().Equal("mock-uuid", result.Id)
}

func (s *UuidsProviderWithCacheSuite) TestUnknownCachedUsername() {
	s.Storage.On("GetUuidForMojangUsername", "username").Return("", "UserName", nil)

	result, err := s.Provider.GetUuid("username")

	s.Require().NoError(err)
	s.Require().Nil(result)
}

func (s *UuidsProviderWithCacheSuite) TestErrorDuringCacheQuery() {
	expectedError := errors.New("mock error")
	s.Storage.On("GetUuidForMojangUsername", "username").Return("", "", expectedError)

	result, err := s.Provider.GetUuid("username")

	s.Require().Same(expectedError, err)
	s.Require().Nil(result)
}

func (s *UuidsProviderWithCacheSuite) TestErrorFromOriginalProvider() {
	expectedError := errors.New("mock error")
	s.Storage.On("GetUuidForMojangUsername", "username").Return("", "", nil)

	s.Original.On("GetUuid", "username").Once().Return(nil, expectedError)

	result, err := s.Provider.GetUuid("username")

	s.Require().Same(expectedError, err)
	s.Require().Nil(result)
}

func TestUuidsProviderWithCache(t *testing.T) {
	suite.Run(t, new(UuidsProviderWithCacheSuite))
}
