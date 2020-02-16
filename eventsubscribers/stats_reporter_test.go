package eventsubscribers

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/elyby/chrly/dispatcher"
	"github.com/elyby/chrly/tests"
)

type StatsReporterTestCase struct {
	Topic         string
	Args          []interface{}
	ExpectedCalls [][]interface{}
}

var statsReporterTestCases = []*StatsReporterTestCase{
	// Before request
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("GET", "http://localhost/skins/username", nil)},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.skins.request", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("GET", "http://localhost/skins?name=username", nil)},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.skins.get_request", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("GET", "http://localhost/cloaks/username", nil)},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.capes.request", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("GET", "http://localhost/cloaks?name=username", nil)},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.capes.get_request", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("GET", "http://localhost/textures/username", nil)},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.textures.request", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("GET", "http://localhost/textures/signed/username", nil)},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.signed_textures.request", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("POST", "http://localhost/api/skins", nil)},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.post.request", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("DELETE", "http://localhost/api/skins/username", nil)},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.request", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("DELETE", "http://localhost/api/skins/id:1", nil)},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.request", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:before_request",
		Args:          []interface{}{httptest.NewRequest("GET", "http://localhost/unknown", nil)},
		ExpectedCalls: nil,
	},
	// After request
	{
		Topic:         "skinsystem:after_request",
		Args:          []interface{}{httptest.NewRequest("POST", "http://localhost/api/skins", nil), 201},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.post.success", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:after_request",
		Args:          []interface{}{httptest.NewRequest("POST", "http://localhost/api/skins", nil), 400},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.post.validation_failed", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:after_request",
		Args:          []interface{}{httptest.NewRequest("DELETE", "http://localhost/api/skins/username", nil), 204},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.success", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:after_request",
		Args:          []interface{}{httptest.NewRequest("DELETE", "http://localhost/api/skins/username", nil), 404},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.not_found", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:after_request",
		Args:          []interface{}{httptest.NewRequest("DELETE", "http://localhost/api/skins/id:1", nil), 204},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.success", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:after_request",
		Args:          []interface{}{httptest.NewRequest("DELETE", "http://localhost/api/skins/id:1", nil), 404},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.not_found", int64(1)},
		},
	},
	{
		Topic:         "skinsystem:after_request",
		Args:          []interface{}{httptest.NewRequest("DELETE", "http://localhost/unknown", nil), 404},
		ExpectedCalls: nil,
	},
	// Authenticator
	{
		Topic:         "authenticator:success",
		Args:          []interface{}{},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.authentication.challenge", int64(1)},
			{"IncCounter", "mock_prefix.authentication.success", int64(1)},
		},
	},
	{
		Topic:         "authentication:error",
		Args:          []interface{}{errors.New("error")},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.authentication.challenge", int64(1)},
			{"IncCounter", "mock_prefix.authentication.failed", int64(1)},
		},
	},
}

func TestStatsReporter_handleHTTPRequest(t *testing.T) {
	for _, c := range statsReporterTestCases {
		t.Run(c.Topic, func(t *testing.T) {
			wdMock := &tests.WdMock{}
			if c.ExpectedCalls != nil {
				for _, c := range c.ExpectedCalls {
					topicName, _ := c[0].(string)
					wdMock.On(topicName, c[1:]...).Once()
				}
			}

			reporter := &StatsReporter{
				Reporter: wdMock,
				Prefix:   "mock_prefix",
			}

			d := dispatcher.New()
			reporter.ConfigureWithDispatcher(d)
			d.Emit(c.Topic, c.Args...)

			wdMock.AssertExpectations(t)
		})
	}
}
