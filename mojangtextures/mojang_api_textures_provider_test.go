package mojangtextures

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/api/mojang"
)

type mojangUuidToTexturesRequestMock struct {
	mock.Mock
}

func (o *mojangUuidToTexturesRequestMock) UuidToTextures(uuid string, signed bool) (*mojang.SignedTexturesResponse, error) {
	args := o.Called(uuid, signed)
	var result *mojang.SignedTexturesResponse
	if casted, ok := args.Get(0).(*mojang.SignedTexturesResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

type mojangApiTexturesProviderTestSuite struct {
	suite.Suite

	Provider  *MojangApiTexturesProvider
	Emitter   *mockEmitter
	MojangApi *mojangUuidToTexturesRequestMock
}

func (suite *mojangApiTexturesProviderTestSuite) SetupTest() {
	suite.Emitter = &mockEmitter{}
	suite.MojangApi = &mojangUuidToTexturesRequestMock{}

	suite.Provider = &MojangApiTexturesProvider{
		Emitter: suite.Emitter,
	}

	uuidToTextures = suite.MojangApi.UuidToTextures
}

func (suite *mojangApiTexturesProviderTestSuite) TearDownTest() {
	suite.MojangApi.AssertExpectations(suite.T())
	suite.Emitter.AssertExpectations(suite.T())
}

func TestMojangApiTexturesProvider(t *testing.T) {
	suite.Run(t, new(mojangApiTexturesProviderTestSuite))
}

func (suite *mojangApiTexturesProviderTestSuite) TestGetTextures() {
	expectedResult := &mojang.SignedTexturesResponse{
		Id:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "username",
	}
	suite.MojangApi.On("UuidToTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true).Once().Return(expectedResult, nil)

	suite.Emitter.On("Emit",
		"mojang_textures:mojang_api_textures_provider:before_request",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	).Once()
	suite.Emitter.On("Emit",
		"mojang_textures:mojang_api_textures_provider:after_request",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		expectedResult,
		nil,
	).Once()

	result, err := suite.Provider.GetTextures("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	suite.Assert().Equal(expectedResult, result)
	suite.Assert().Nil(err)
}

func (suite *mojangApiTexturesProviderTestSuite) TestGetTexturesWithError() {
	var expectedResponse *mojang.SignedTexturesResponse
	expectedError := &mojang.TooManyRequestsError{}
	suite.MojangApi.On("UuidToTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true).Once().Return(nil, expectedError)

	suite.Emitter.On("Emit",
		"mojang_textures:mojang_api_textures_provider:before_request",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	).Once()
	suite.Emitter.On("Emit",
		"mojang_textures:mojang_api_textures_provider:after_request",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		expectedResponse,
		expectedError,
	).Once()

	result, err := suite.Provider.GetTextures("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	suite.Assert().Nil(result)
	suite.Assert().Equal(expectedError, err)
}
