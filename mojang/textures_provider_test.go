package mojang

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var signedTexturesResponse = &SignedTexturesResponse{
	Id:   "dead24f9a4fa4877b7b04c8c6c72bb46",
	Name: "mock",
	Props: []*Property{
		{
			Name: "textures",
			Value: EncodeTextures(&TexturesProp{
				Timestamp:   time.Now().UnixNano() / 10e5,
				ProfileID:   "dead24f9a4fa4877b7b04c8c6c72bb46",
				ProfileName: "mock",
				Textures: &TexturesResponse{
					Skin: &SkinTexturesResponse{
						Url: "http://textures.minecraft.net/texture/74d1e08b0bb7e9f590af27758125bbed1778ac6cef729aedfcb9613e9911ae75",
					},
				},
			}),
		},
	},
}

type MojangUuidToTexturesRequestMock struct {
	mock.Mock
}

func (m *MojangUuidToTexturesRequestMock) UuidToTextures(uuid string, signed bool) (*SignedTexturesResponse, error) {
	args := m.Called(uuid, signed)
	var result *SignedTexturesResponse
	if casted, ok := args.Get(0).(*SignedTexturesResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

type MojangApiTexturesProviderSuite struct {
	suite.Suite

	Provider  *MojangApiTexturesProvider
	MojangApi *MojangUuidToTexturesRequestMock
}

func (s *MojangApiTexturesProviderSuite) SetupTest() {
	s.MojangApi = &MojangUuidToTexturesRequestMock{}
	s.Provider = &MojangApiTexturesProvider{
		MojangApiTexturesEndpoint: s.MojangApi.UuidToTextures,
	}
}

func (s *MojangApiTexturesProviderSuite) TearDownTest() {
	s.MojangApi.AssertExpectations(s.T())
}

func (s *MojangApiTexturesProviderSuite) TestGetTextures() {
	s.MojangApi.On("UuidToTextures", "dead24f9a4fa4877b7b04c8c6c72bb46", true).Once().Return(signedTexturesResponse, nil)

	result, err := s.Provider.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

	s.Require().NoError(err)
	s.Require().Equal(signedTexturesResponse, result)
}

func (s *MojangApiTexturesProviderSuite) TestGetTexturesWithError() {
	expectedError := errors.New("mock error")
	s.MojangApi.On("UuidToTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true).Once().Return(nil, expectedError)

	result, err := s.Provider.GetTextures("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	s.Require().Nil(result)
	s.Require().Equal(expectedError, err)
}

func TestMojangApiTexturesProvider(t *testing.T) {
	suite.Run(t, new(MojangApiTexturesProviderSuite))
}

type TexturesProviderWithInMemoryCacheSuite struct {
	suite.Suite
	Original *TexturesProviderMock
	Provider *TexturesProviderWithInMemoryCache
}

func (s *TexturesProviderWithInMemoryCacheSuite) SetupTest() {
	s.Original = &TexturesProviderMock{}
	s.Provider = NewTexturesProviderWithInMemoryCache(s.Original)
}

func (s *TexturesProviderWithInMemoryCacheSuite) TearDownTest() {
	s.Original.AssertExpectations(s.T())
	s.Provider.StopGC()
}

func (s *TexturesProviderWithInMemoryCacheSuite) TestGetTexturesWithSuccessfulOriginalProviderResponse() {
	s.Original.On("GetTextures", "uuid").Once().Return(signedTexturesResponse, nil)
	// Do the call multiple times to ensure, that there will be only one call to the Original provider
	for i := 0; i < 5; i++ {
		result, err := s.Provider.GetTextures("uuid")

		s.Require().NoError(err)
		s.Require().Same(signedTexturesResponse, result)
	}
}

func (s *TexturesProviderWithInMemoryCacheSuite) TestGetTexturesWithEmptyOriginalProviderResponse() {
	s.Original.On("GetTextures", "uuid").Once().Return(nil, nil)
	// Do the call multiple times to ensure, that there will be only one call to the original provider
	for i := 0; i < 5; i++ {
		result, err := s.Provider.GetTextures("uuid")

		s.Require().NoError(err)
		s.Require().Nil(result)
	}
}

func (s *TexturesProviderWithInMemoryCacheSuite) TestGetTexturesWithErrorFromOriginalProvider() {
	expectedErr := errors.New("mock error")
	s.Original.On("GetTextures", "uuid").Times(5).Return(nil, expectedErr)
	// Do the call multiple times to ensure, that the error will not be cached and there will be a request on each call
	for i := 0; i < 5; i++ {
		result, err := s.Provider.GetTextures("uuid")

		s.Require().Same(expectedErr, err)
		s.Require().Nil(result)
	}
}

func TestTexturesProviderWithInMemoryCache(t *testing.T) {
	suite.Run(t, new(TexturesProviderWithInMemoryCacheSuite))
}
