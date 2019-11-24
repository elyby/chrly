package mojangtextures

import (
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	mocks "github.com/elyby/chrly/tests"
)

type remoteApiUuidsProviderTestSuite struct {
	suite.Suite

	Provider *RemoteApiUuidsProvider
	Logger   *mocks.WdMock
}

func (suite *remoteApiUuidsProviderTestSuite) SetupSuite() {
	client := &http.Client{}
	gock.InterceptClient(client)

	HttpClient = client
}

func (suite *remoteApiUuidsProviderTestSuite) SetupTest() {
	suite.Logger = &mocks.WdMock{}
	suite.Provider = &RemoteApiUuidsProvider{
		Logger: suite.Logger,
	}
}

func (suite *remoteApiUuidsProviderTestSuite) TearDownTest() {
	suite.Logger.AssertExpectations(suite.T())
	gock.Off()
}

func TestRemoteApiUuidsProvider(t *testing.T) {
	suite.Run(t, new(remoteApiUuidsProviderTestSuite))
}

func (suite *remoteApiUuidsProviderTestSuite) TestGetUuidForValidUsername() {
	suite.Logger.On("IncCounter", "mojang_textures.usernames.request", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.request_time", mock.Anything).Once()

	gock.New("http://example.com").
		Get("/subpath/username").
		Reply(200).
		JSON(map[string]interface{}{
			"id":   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"name": "username",
		})

	suite.Provider.Url = shouldParseUrl("http://example.com/subpath")
	result, err := suite.Provider.GetUuid("username")

	assert := suite.Assert()
	if assert.NoError(err) {
		assert.Equal("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", result.Id)
		assert.Equal("username", result.Name)
		assert.False(result.IsLegacy)
		assert.False(result.IsDemo)
	}
}

func (suite *remoteApiUuidsProviderTestSuite) TestGetUuidForNotExistsUsername() {
	suite.Logger.On("IncCounter", "mojang_textures.usernames.request", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.request_time", mock.Anything).Once()

	gock.New("http://example.com").
		Get("/subpath/username").
		Reply(204)

	suite.Provider.Url = shouldParseUrl("http://example.com/subpath")
	result, err := suite.Provider.GetUuid("username")

	assert := suite.Assert()
	assert.Nil(result)
	assert.Nil(err)
}

func (suite *remoteApiUuidsProviderTestSuite) TestGetUuidForNon20xResponse() {
	suite.Logger.On("IncCounter", "mojang_textures.usernames.request", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.request_time", mock.Anything).Once()

	gock.New("http://example.com").
		Get("/subpath/username").
		Reply(504).
		BodyString("504 Gateway Timeout")

	suite.Provider.Url = shouldParseUrl("http://example.com/subpath")
	result, err := suite.Provider.GetUuid("username")

	assert := suite.Assert()
	assert.Nil(result)
	assert.EqualError(err, "Unexpected remote api response")
}

func (suite *remoteApiUuidsProviderTestSuite) TestGetUuidForNotSuccessRequest() {
	suite.Logger.On("IncCounter", "mojang_textures.usernames.request", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.request_time", mock.Anything).Once()

	expectedError := &net.OpError{Op: "dial"}

	gock.New("http://example.com").
		Get("/subpath/username").
		ReplyError(expectedError)

	suite.Provider.Url = shouldParseUrl("http://example.com/subpath")
	result, err := suite.Provider.GetUuid("username")

	assert := suite.Assert()
	assert.Nil(result)
	if assert.Error(err) {
		assert.IsType(&url.Error{}, err)
		casterErr, _ := err.(*url.Error)
		assert.Equal(expectedError, casterErr.Err)
	}
}

func (suite *remoteApiUuidsProviderTestSuite) TestGetUuidForInvalidSuccessResponse() {
	suite.Logger.On("IncCounter", "mojang_textures.usernames.request", int64(1)).Once()
	suite.Logger.On("RecordTimer", "mojang_textures.usernames.request_time", mock.Anything).Once()

	gock.New("http://example.com").
		Get("/subpath/username").
		Reply(200).
		BodyString("completely not json")

	suite.Provider.Url = shouldParseUrl("http://example.com/subpath")
	result, err := suite.Provider.GetUuid("username")

	assert := suite.Assert()
	assert.Nil(result)
	assert.Error(err)
}

func shouldParseUrl(rawUrl string) url.URL {
	url, err := url.Parse(rawUrl)
	if err != nil {
		panic(err)
	}

	return *url
}
