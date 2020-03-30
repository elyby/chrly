package eventsubscribers

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/dispatcher"
	"github.com/elyby/chrly/tests"
)

type StatsReporterTestCase struct {
	Events        map[string][]interface{}
	ExpectedCalls [][]interface{}
}

var statsReporterTestCases = []*StatsReporterTestCase{
	// Before request
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("GET", "http://localhost/skins/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.skins.request", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("GET", "http://localhost/skins?name=username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.skins.get_request", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("GET", "http://localhost/cloaks/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.capes.request", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("GET", "http://localhost/cloaks?name=username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.capes.get_request", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("GET", "http://localhost/textures/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.textures.request", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("GET", "http://localhost/textures/signed/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.signed_textures.request", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("POST", "http://localhost/api/skins", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.post.request", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("DELETE", "http://localhost/api/skins/username", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.request", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("DELETE", "http://localhost/api/skins/id:1", nil)},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.request", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:before_request": {httptest.NewRequest("GET", "http://localhost/unknown", nil)},
		},
		ExpectedCalls: nil,
	},
	// After request
	{
		Events: map[string][]interface{}{
			"skinsystem:after_request": {httptest.NewRequest("POST", "http://localhost/api/skins", nil), 201},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.post.success", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:after_request": {httptest.NewRequest("POST", "http://localhost/api/skins", nil), 400},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.post.validation_failed", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:after_request": {httptest.NewRequest("DELETE", "http://localhost/api/skins/username", nil), 204},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.success", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:after_request": {httptest.NewRequest("DELETE", "http://localhost/api/skins/username", nil), 404},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.not_found", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:after_request": {httptest.NewRequest("DELETE", "http://localhost/api/skins/id:1", nil), 204},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.success", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:after_request": {httptest.NewRequest("DELETE", "http://localhost/api/skins/id:1", nil), 404},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.api.skins.delete.not_found", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"skinsystem:after_request": {httptest.NewRequest("DELETE", "http://localhost/unknown", nil), 404},
		},
		ExpectedCalls: nil,
	},
	// Authenticator
	{
		Events: map[string][]interface{}{
			"authenticator:success": {},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.authentication.challenge", int64(1)},
			{"IncCounter", "mock_prefix.authentication.success", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"authentication:error": {errors.New("error")},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.authentication.challenge", int64(1)},
			{"IncCounter", "mock_prefix.authentication.failed", int64(1)},
		},
	},
	// Mojang signed textures provider
	{
		Events: map[string][]interface{}{
			"mojang_textures:before_result": {"username", ""},
			"mojang_textures:after_result":  {"username", &mojang.SignedTexturesResponse{}, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"RecordTimer", "mock_prefix.mojang_textures.result_time", mock.AnythingOfType("time.Duration")},
		},
	},
	{
		Events: map[string][]interface{}{
			"mojang_textures:textures:before_call": {"аааааааааааааааааааааааааааааааа"},
			"mojang_textures:textures:after_call":  {"аааааааааааааааааааааааааааааааа", &mojang.SignedTexturesResponse{}, nil},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.mojang_textures.textures.request", int64(1)},
			{"IncCounter", "mock_prefix.mojang_textures.usernames.textures_hit", int64(1)},
			{"RecordTimer", "mock_prefix.mojang_textures.textures.request_time", mock.AnythingOfType("time.Duration")},
		},
	},
	// Batch UUIDs provider
	{
		Events: map[string][]interface{}{
			"mojang_textures:batch_uuids_provider:queued": {"username"},
		},
		ExpectedCalls: [][]interface{}{
			{"IncCounter", "mock_prefix.mojang_textures.usernames.queued", int64(1)},
		},
	},
	{
		Events: map[string][]interface{}{
			"mojang_textures:batch_uuids_provider:round": {[]string{"username1", "username2"}, 5},
		},
		ExpectedCalls: [][]interface{}{
			{"UpdateGauge", "mock_prefix.mojang_textures.usernames.iteration_size", int64(2)},
			{"UpdateGauge", "mock_prefix.mojang_textures.usernames.queue_size", int64(5)},
		},
	},
	{
		Events: map[string][]interface{}{
			"mojang_textures:batch_uuids_provider:before_round": {},
			"mojang_textures:batch_uuids_provider:after_round":  {},
		},
		ExpectedCalls: [][]interface{}{
			{"RecordTimer", "mock_prefix.mojang_textures.usernames.round_time", mock.AnythingOfType("time.Duration")},
		},
	},
}

func TestStatsReporter_handleEvents(t *testing.T) {
	for _, c := range statsReporterTestCases {
		t.Run("handle events", func(t *testing.T) {
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
			for event, args := range c.Events {
				d.Emit(event, args...)
			}

			wdMock.AssertExpectations(t)
		})
	}
}
