package mojangtextures

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/api/mojang"
	mocks "github.com/elyby/chrly/tests"
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
	Logger    *mocks.WdMock
	MojangApi *mojangUuidToTexturesRequestMock
}

func (suite *mojangApiTexturesProviderTestSuite) SetupTest() {
	suite.Logger = &mocks.WdMock{}
	suite.MojangApi = &mojangUuidToTexturesRequestMock{}

	suite.Provider = &MojangApiTexturesProvider{
		Logger: suite.Logger,
	}

	uuidToTextures = suite.MojangApi.UuidToTextures
}

func (suite *mojangApiTexturesProviderTestSuite) TearDownTest() {
	suite.MojangApi.AssertExpectations(suite.T())
	suite.Logger.AssertExpectations(suite.T())
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

	suite.Logger.On("IncCounter", "mojang_textures.textures.request", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.textures.request_time", mock.Anything).Once()

	result, err := suite.Provider.GetTextures("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	suite.Assert().Equal(expectedResult, result)
	suite.Assert().Nil(err)
}

func (suite *mojangApiTexturesProviderTestSuite) TestGetTexturesWithError() {
	expectedError := &mojang.TooManyRequestsError{}
	suite.MojangApi.On("UuidToTextures", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true).Once().Return(nil, expectedError)

	suite.Logger.On("IncCounter", "mojang_textures.textures.request", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.textures.request_time", mock.Anything).Once()

	result, err := suite.Provider.GetTextures("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	suite.Assert().Nil(result)
	suite.Assert().Equal(expectedError, err)
}
