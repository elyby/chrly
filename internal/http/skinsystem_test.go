package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	testify "github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"ely.by/chrly/internal/db"
)

type ProfilesProviderMock struct {
	mock.Mock
}

func (m *ProfilesProviderMock) FindProfileByUsername(ctx context.Context, username string, allowProxy bool) (*db.Profile, error) {
	args := m.Called(ctx, username, allowProxy)
	var result *db.Profile
	if casted, ok := args.Get(0).(*db.Profile); ok {
		result = casted
	}

	return result, args.Error(1)
}

type SkinsystemTestSuite struct {
	suite.Suite

	App *Skinsystem

	ProfilesProvider *ProfilesProviderMock
}

/********************
 * Setup test suite *
 ********************/

func (t *SkinsystemTestSuite) SetupSubTest() {
	t.ProfilesProvider = &ProfilesProviderMock{}

	t.App, _ = NewSkinsystemApi(
		t.ProfilesProvider,
		"texturesParamName",
		"texturesParamValue",
	)
}

func (t *SkinsystemTestSuite) TearDownSubTest() {
	t.ProfilesProvider.AssertExpectations(t.T())
}

func (t *SkinsystemTestSuite) TestSkinHandler() {
	for _, url := range []string{"http://chrly/skins/mock_username", "http://chrly/skins?name=mock_username"} {
		t.Run("known username with a skin", func() {
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// TODO: see the TODO about context above
			t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{
				SkinUrl: "https://example.com/skin.png",
			}, nil)

			t.App.Handler().ServeHTTP(w, req)

			result := w.Result()
			t.Equal(http.StatusMovedPermanently, result.StatusCode)
			t.Equal("https://example.com/skin.png", result.Header.Get("Location"))
		})

		t.Run("known username without a skin", func() {
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{}, nil)

			t.App.Handler().ServeHTTP(w, req)

			result := w.Result()
			t.Equal(http.StatusNotFound, result.StatusCode)
		})

		t.Run("err from profiles provider", func() {
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(nil, errors.New("mock error"))

			t.App.Handler().ServeHTTP(w, req)

			result := w.Result()
			t.Equal(http.StatusInternalServerError, result.StatusCode)
		})
	}

	t.Run("username with png extension", func() {
		req := httptest.NewRequest("GET", "http://chrly/skins/mock_username.png", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{
			SkinUrl: "https://example.com/skin.png",
		}, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusMovedPermanently, result.StatusCode)
		t.Equal("https://example.com/skin.png", result.Header.Get("Location"))
	})

	t.Run("no name param", func() {
		req := httptest.NewRequest("GET", "http://chrly/skins", nil)
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		t.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func (t *SkinsystemTestSuite) TestCapeHandler() {
	for _, url := range []string{"http://chrly/cloaks/mock_username", "http://chrly/cloaks?name=mock_username"} {
		t.Run("known username with a skin", func() {
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// TODO: I can't find a way to verify that it's the context from the request that was passed in,
			//       as the Mux calls WithValue() on it, which creates a new Context and I haven't been able
			//       to find a way to verify that the passed context matches the base
			t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{
				CapeUrl: "https://example.com/cape.png",
			}, nil)

			t.App.Handler().ServeHTTP(w, req)

			result := w.Result()
			t.Equal(http.StatusMovedPermanently, result.StatusCode)
			t.Equal("https://example.com/cape.png", result.Header.Get("Location"))
		})

		t.Run("known username without a skin", func() {
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{}, nil)

			t.App.Handler().ServeHTTP(w, req)

			result := w.Result()
			t.Equal(http.StatusNotFound, result.StatusCode)
		})

		t.Run("err from profiles provider", func() {
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(nil, errors.New("mock error"))

			t.App.Handler().ServeHTTP(w, req)

			result := w.Result()
			t.Equal(http.StatusInternalServerError, result.StatusCode)
		})
	}

	t.Run("username with png extension", func() {
		req := httptest.NewRequest("GET", "http://chrly/cloaks/mock_username.png", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{
			CapeUrl: "https://example.com/cape.png",
		}, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusMovedPermanently, result.StatusCode)
		t.Equal("https://example.com/cape.png", result.Header.Get("Location"))
	})

	t.Run("no name param", func() {
		req := httptest.NewRequest("GET", "http://chrly/cloaks", nil)
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)

		resp := w.Result()
		t.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func (t *SkinsystemTestSuite) TestTexturesHandler() {
	t.Run("known username with both textures", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
		w := httptest.NewRecorder()

		// TODO: see the TODO about context above
		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{
			SkinUrl: "https://example.com/skin.png",
			CapeUrl: "https://example.com/cape.png",
		}, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusOK, result.StatusCode)
		t.Equal("application/json", result.Header.Get("Content-Type"))
		body, _ := io.ReadAll(result.Body)
		t.JSONEq(`{
			"SKIN": {
				"url": "https://example.com/skin.png"
			},
			"CAPE": {
				"url": "https://example.com/cape.png"
			}
		}`, string(body))
	})

	t.Run("known username with only slim skin", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{
			SkinUrl:   "https://example.com/skin.png",
			SkinModel: "slim",
		}, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		body, _ := io.ReadAll(result.Body)
		t.JSONEq(`{
			"SKIN": {
				"url": "https://example.com/skin.png",
				"metadata": {
					"model": "slim"
				}
			}
		}`, string(body))
	})

	t.Run("known username with only cape", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{
			CapeUrl: "https://example.com/cape.png",
		}, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		body, _ := io.ReadAll(result.Body)
		t.JSONEq(`{
			"CAPE": {
				"url": "https://example.com/cape.png"
			}
		}`, string(body))
	})

	t.Run("known username without any textures", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{}, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusNoContent, result.StatusCode)
		body, _ := io.ReadAll(result.Body)
		t.Empty(body)
	})

	t.Run("unknown username", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(nil, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusNotFound, result.StatusCode)
		body, _ := io.ReadAll(result.Body)
		t.Empty(body)
	})

	t.Run("err from profiles provider", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(nil, errors.New("mock error"))

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusInternalServerError, result.StatusCode)
	})
}

func (t *SkinsystemTestSuite) TestSignedTextures() {
	t.Run("exists profile with mojang textures", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/signed/mock_username", nil)
		w := httptest.NewRecorder()

		// TODO: see the TODO about context above
		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", false).Return(&db.Profile{
			Uuid:            "mock-uuid",
			Username:        "mock",
			MojangTextures:  "mock-mojang-textures",
			MojangSignature: "mock-mojang-signature",
		}, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusOK, result.StatusCode)
		t.Equal("application/json", result.Header.Get("Content-Type"))
		body, _ := io.ReadAll(result.Body)
		t.JSONEq(`{
			"id": "mock-uuid",
			"name": "mock",
			"properties": [
				{
					"name": "textures",
					"signature": "mock-mojang-signature",
					"value": "mock-mojang-textures"
				},
				{
					"name": "texturesParamName",
					"value": "texturesParamValue"
				}
			]
		}`, string(body))
	})

	t.Run("exists profile without mojang textures", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/signed/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", false).Return(&db.Profile{}, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusNoContent, result.StatusCode)
		body, _ := io.ReadAll(result.Body)
		t.Empty(body)
	})

	t.Run("not exists profile", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/signed/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", false).Return(nil, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusNotFound, result.StatusCode)
		body, _ := io.ReadAll(result.Body)
		t.Empty(body)
	})

	t.Run("err from profiles provider", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/signed/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", false).Return(nil, errors.New("mock error"))

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusInternalServerError, result.StatusCode)
	})

	t.Run("should allow proxying when specified get param", func() {
		req := httptest.NewRequest("GET", "http://chrly/textures/signed/mock_username?proxy=true", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(nil, nil)

		t.App.Handler().ServeHTTP(w, req)
	})
}

func TestSkinsystem(t *testing.T) {
	suite.Run(t, new(SkinsystemTestSuite))
}

func TestParseUsername(t *testing.T) {
	assert := testify.New(t)
	assert.Equal("test", parseUsername("test.png"), "Function should trim .png at end")
	assert.Equal("test", parseUsername("test"), "Function should return string itself, if it not contains .png at end")
}
