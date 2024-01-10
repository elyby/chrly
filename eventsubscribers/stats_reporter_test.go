package eventsubscribers

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mono83/slf"

	"github.com/elyby/chrly/dispatcher"

	"github.com/stretchr/testify/mock"
)

func prepareStatsReporterArgs(name string, value interface{}, params []slf.Param) []interface{} {
	args := []interface{}{name, value}
	for _, v := range params {
		args = append(args, v.(interface{}))
	}

	return args
}

type StatsReporterMock struct {
	mock.Mock
}

func (r *StatsReporterMock) IncCounter(name string, value int64, params ...slf.Param) {
	r.Called(prepareStatsReporterArgs(name, value, params)...)
}

func (r *StatsReporterMock) UpdateGauge(name string, value int64, params ...slf.Param) {
	r.Called(prepareStatsReporterArgs(name, value, params)...)
}

func (r *StatsReporterMock) RecordTimer(name string, duration time.Duration, params ...slf.Param) {
	r.Called(prepareStatsReporterArgs(name, duration, params)...)
}

func (r *StatsReporterMock) Timer(name string, params ...slf.Param) slf.Timer {
	return slf.NewTimer(name, params, r)
}

type StatsReporterTestCase struct {
	Events        [][]interface{}
	ExpectedCalls [][]interface{}
}

var statsReporterTestCases = []*StatsReporterTestCase{
	// Before request
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("GET", "http://localhost/skins/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "skins.request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("GET", "http://localhost/skins?name=username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "skins.get_request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("GET", "http://localhost/cloaks/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "capes.request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("GET", "http://localhost/cloaks?name=username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "capes.get_request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("GET", "http://localhost/textures/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "textures.request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("GET", "http://localhost/textures/signed/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "signed_textures.request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("GET", "http://localhost/profile/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "profiles.request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("POST", "http://localhost/api/skins", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "api.skins.post.request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("DELETE", "http://localhost/api/skins/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "api.skins.delete.request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("DELETE", "http://localhost/api/skins/id:1", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "api.skins.delete.request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:before_request", httptest.NewRequest("GET", "http://localhost/unknown", nil)},
		},
		ExpectedCalls: nil,
	},
	// After request
	{
		Events: [][]interface{}{
			{"skinsystem:after_request", httptest.NewRequest("POST", "http://localhost/api/skins", nil), 201},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "api.skins.post.success", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:after_request", httptest.NewRequest("POST", "http://localhost/api/skins", nil), 400},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "api.skins.post.validation_failed", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:after_request", httptest.NewRequest("DELETE", "http://localhost/api/skins/username", nil), 204},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "api.skins.delete.success", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:after_request", httptest.NewRequest("DELETE", "http://localhost/api/skins/username", nil), 404},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "api.skins.delete.not_found", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:after_request", httptest.NewRequest("DELETE", "http://localhost/api/skins/id:1", nil), 204},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "api.skins.delete.success", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:after_request", httptest.NewRequest("DELETE", "http://localhost/api/skins/id:1", nil), 404},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "api.skins.delete.not_found", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"skinsystem:after_request", httptest.NewRequest("DELETE", "http://localhost/unknown", nil), 404},
		},
		ExpectedCalls: nil,
	},
	// Authenticator
	{
		Events: [][]interface{}{
			{"authenticator:success"},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "authentication.challenge", int64(1)},
			{"IncCounter", "authentication.success", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"authentication:error", errors.New("error")},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "authentication.challenge", int64(1)},
			{"IncCounter", "authentication.failed", int64(1)},
		},
	},
}

func TestStatsReporter(t *testing.T) {
	for _, c := range statsReporterTestCases {
		t.Run("handle events", func(t *testing.T) {
			statsReporterMock := &StatsReporterMock{}
			if c.ExpectedCalls != nil {
				for _, c := range c.ExpectedCalls {
					topicName, _ := c[0].(string)
					statsReporterMock.On(topicName, c[1:]...).Once()
				}
			}

			reporter := &StatsReporter{
				StatsReporter: statsReporterMock,
				Prefix:        "mock_prefix",
			}

			d := dispatcher.New()
			reporter.ConfigureWithDispatcher(d)
			for _, e := range c.Events {
				eventName, _ := e[0].(string)
				d.Emit(eventName, e[1:]...)
			}

			statsReporterMock.AssertExpectations(t)
		})
	}
}
