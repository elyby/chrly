package eventsubscribers

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mono83/slf"

	"github.com/elyby/chrly/api/mojang"
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
	// Mojang signed textures provider
	{
		Events: [][]interface{}{
			{"mojang_textures:call", "username"},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.request", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:usernames:after_cache", "username", "", false, errors.New("error")},
		},
		ExpectedCalls: [][]interface{}{},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:usernames:after_cache", "username", "", false, nil},
		},
		ExpectedCalls: [][]interface{}{},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:usernames:after_cache", "username", "", true, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.usernames.cache_hit_nil", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:usernames:after_cache", "username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.usernames.cache_hit", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:textures:after_cache", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil, errors.New("error")},
		},
		ExpectedCalls: [][]interface{}{},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:textures:after_cache", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil, nil},
		},
		ExpectedCalls: [][]interface{}{},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:textures:after_cache", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", &mojang.SignedTexturesResponse{}, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.textures.cache_hit", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:already_processing", "username"},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.already_scheduled", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:usernames:after_call", "username", nil, errors.New("error")},
		},
		ExpectedCalls: [][]interface{}{},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:usernames:after_call", "username", nil, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.usernames.uuid_miss", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:usernames:after_call", "username", &mojang.ProfileInfo{}, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.usernames.uuid_hit", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:textures:after_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil, errors.New("error")},
		},
		ExpectedCalls: [][]interface{}{},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:textures:after_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.usernames.textures_miss", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:textures:after_call", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", &mojang.SignedTexturesResponse{}, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.usernames.textures_hit", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:before_result", "username", ""},
			{"mojang_textures:after_result", "username", &mojang.SignedTexturesResponse{}, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"RecordTimer", "mojang_textures.result_time", mock.AnythingOfType("time.Duration")},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:textures:before_call", "аааааааааааааааааааааааааааааааа"},
			{"mojang_textures:textures:after_call", "аааааааааааааааааааааааааааааааа", &mojang.SignedTexturesResponse{}, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.textures.request", int64(1)},
			{"IncCounter", "mojang_textures.usernames.textures_hit", int64(1)},
			{"RecordTimer", "mojang_textures.textures.request_time", mock.AnythingOfType("time.Duration")},
		},
	},
	// Batch UUIDs provider
	{
		Events: [][]interface{}{
			{"mojang_textures:batch_uuids_provider:queued", "username"},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mojang_textures.usernames.queued", int64(1)},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:batch_uuids_provider:round", []string{"username1", "username2"}, 5},
			{"mojang_textures:batch_uuids_provider:result", []string{"username1", "username2"}, []*mojang.ProfileInfo{}, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"UpdateGauge", "mojang_textures.usernames.iteration_size", int64(2)},
			{"UpdateGauge", "mojang_textures.usernames.queue_size", int64(5)},
			{"RecordTimer", "mojang_textures.usernames.round_time", mock.AnythingOfType("time.Duration")},
		},
	},
	{
		Events: [][]interface{}{
			{"mojang_textures:batch_uuids_provider:round", []string{}, 0},
			// This event will be not emitted, but we emit it to ensure, that RecordTimer will not be called
			{"mojang_textures:batch_uuids_provider:result", []string{}, []*mojang.ProfileInfo{}, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"UpdateGauge", "mojang_textures.usernames.iteration_size", int64(0)},
			{"UpdateGauge", "mojang_textures.usernames.queue_size", int64(0)},
			// Should not call RecordTimer
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

type redisPoolCheckableMock struct {
	mock.Mock
}

func (r *redisPoolCheckableMock) Avail() int {
	return r.Called().Int(0)
}

func TestAvailableRedisPoolSizeReporter(t *testing.T) {
	poolMock := &redisPoolCheckableMock{}
	poolMock.On("Avail").Return(5).Times(3)
	reporterMock := &StatsReporterMock{}
	reporterMock.On("UpdateGauge", "redis.pool.available", int64(5)).Times(3)

	ctx, cancel := context.WithCancel(context.Background())

	creator := AvailableRedisPoolSizeReporter(poolMock, 10*time.Millisecond, ctx)
	creator(reporterMock)

	time.Sleep(35 * time.Millisecond)

	cancel()

	poolMock.AssertExpectations(t)
	reporterMock.AssertExpectations(t)
}
