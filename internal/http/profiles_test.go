package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ely.by/chrly/internal/db"
	"ely.by/chrly/internal/profiles"
)

type ProfilesManagerMock struct {
	mock.Mock
}

func (m *ProfilesManagerMock) PersistProfile(ctx context.Context, profile *db.Profile) error {
	return m.Called(ctx, profile).Error(0)
}

func (m *ProfilesManagerMock) RemoveProfileByUuid(ctx context.Context, uuid string) error {
	return m.Called(ctx, uuid).Error(0)
}

type ProfilesTestSuite struct {
	suite.Suite

	App *ProfilesApi

	ProfilesManager *ProfilesManagerMock
}

func (t *ProfilesTestSuite) SetupSubTest() {
	t.ProfilesManager = &ProfilesManagerMock{}
	t.App, _ = NewProfilesApi(t.ProfilesManager)
}

func (t *ProfilesTestSuite) TearDownSubTest() {
	t.ProfilesManager.AssertExpectations(t.T())
}

func (t *ProfilesTestSuite) TestPostProfile() {
	t.Run("successfully post profile", func() {
		t.ProfilesManager.On("PersistProfile", mock.Anything, &db.Profile{
			Uuid:            "0f657aa8-bfbe-415d-b700-5750090d3af3",
			Username:        "mock_username",
			SkinUrl:         "https://example.com/skin.png",
			SkinModel:       "slim",
			CapeUrl:         "https://example.com/cape.png",
			MojangTextures:  "bW9jawo=",
			MojangSignature: "bW9jawo=",
		}).Once().Return(nil)

		req := httptest.NewRequest("POST", "http://chrly/", bytes.NewBufferString(url.Values{
			"uuid":            {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"username":        {"mock_username"},
			"skinUrl":         {"https://example.com/skin.png"},
			"skinModel":       {"slim"},
			"capeUrl":         {"https://example.com/cape.png"},
			"mojangTextures":  {"bW9jawo="},
			"mojangSignature": {"bW9jawo="},
		}.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)
		result := w.Result()

		t.Equal(http.StatusCreated, result.StatusCode)
		body, _ := io.ReadAll(result.Body)
		t.Empty(body)
	})

	t.Run("handle malformed body", func() {
		req := httptest.NewRequest("POST", "http://chrly/", strings.NewReader("invalid;=url?encoded_string"))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)
		result := w.Result()

		t.Equal(http.StatusBadRequest, result.StatusCode)
		body, _ := io.ReadAll(result.Body)
		t.JSONEq(`{
			"errors": {
				"body": [
					"The body of the request must be a valid url-encoded string"
				]
			}
		}`, string(body))
	})

	t.Run("receive validation errors", func() {
		t.ProfilesManager.On("PersistProfile", mock.Anything, mock.Anything).Once().Return(&profiles.ValidationError{
			Errors: map[string][]string{
				"mock": {"error1", "error2"},
			},
		})

		req := httptest.NewRequest("POST", "http://chrly/", strings.NewReader(""))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)
		result := w.Result()

		t.Equal(http.StatusBadRequest, result.StatusCode)
		body, _ := io.ReadAll(result.Body)
		t.JSONEq(`{
			"errors": {
				"mock": [
					"error1",
					"error2"
				]
			}
		}`, string(body))
	})

	t.Run("receive other error", func() {
		t.ProfilesManager.On("PersistProfile", mock.Anything, mock.Anything).Once().Return(errors.New("mock error"))

		req := httptest.NewRequest("POST", "http://chrly/", strings.NewReader(""))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)
		result := w.Result()

		t.Equal(http.StatusInternalServerError, result.StatusCode)
	})
}

func (t *ProfilesTestSuite) TestDeleteProfileByUuid() {
	t.Run("successfully delete", func() {
		t.ProfilesManager.On("RemoveProfileByUuid", mock.Anything, "0f657aa8-bfbe-415d-b700-5750090d3af3").Once().Return(nil)

		req := httptest.NewRequest("DELETE", "http://chrly/0f657aa8-bfbe-415d-b700-5750090d3af3", nil)
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		t.Equal(http.StatusNoContent, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		t.Empty(body)
	})

	t.Run("error from manager", func() {
		t.ProfilesManager.On("RemoveProfileByUuid", mock.Anything, mock.Anything).Return(errors.New("mock error"))

		req := httptest.NewRequest("DELETE", "http://chrly/0f657aa8-bfbe-415d-b700-5750090d3af3", nil)
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		t.Equal(http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestProfilesApi(t *testing.T) {
	suite.Run(t, new(ProfilesTestSuite))
}
