package http

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

type TexturesSignerMock struct {
	mock.Mock
}

func (m *TexturesSignerMock) SignTextures(textures string) (string, error) {
	args := m.Called(textures)
	return args.String(0), args.Error(1)
}

func (m *TexturesSignerMock) GetPublicKey() (*rsa.PublicKey, error) {
	args := m.Called()
	var publicKey *rsa.PublicKey
	if casted, ok := args.Get(0).(*rsa.PublicKey); ok {
		publicKey = casted
	}

	return publicKey, args.Error(1)
}

type SkinsystemTestSuite struct {
	suite.Suite

	App *Skinsystem

	ProfilesProvider *ProfilesProviderMock
	TexturesSigner   *TexturesSignerMock
}

/********************
 * Setup test suite *
 ********************/

func (t *SkinsystemTestSuite) SetupSubTest() {
	timeNow = func() time.Time {
		CET, _ := time.LoadLocation("CET")
		return time.Date(2021, 02, 25, 01, 50, 23, 0, CET)
	}

	t.ProfilesProvider = &ProfilesProviderMock{}
	t.TexturesSigner = &TexturesSignerMock{}

	t.App = &Skinsystem{
		ProfilesProvider:        t.ProfilesProvider,
		TexturesSigner:          t.TexturesSigner,
		TexturesExtraParamName:  "texturesParamName",
		TexturesExtraParamValue: "texturesParamValue",
	}
}

func (t *SkinsystemTestSuite) TearDownSubTest() {
	t.ProfilesProvider.AssertExpectations(t.T())
	t.TexturesSigner.AssertExpectations(t.T())
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

func (t *SkinsystemTestSuite) TestProfile() {
	t.Run("exists profile with skin and cape", func() {
		req := httptest.NewRequest("GET", "http://chrly/profile/mock_username", nil)
		w := httptest.NewRecorder()

		// TODO: see the TODO about context above
		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{
			Uuid:      "mock-uuid",
			Username:  "mock_username",
			SkinUrl:   "https://example.com/skin.png",
			SkinModel: "slim",
			CapeUrl:   "https://example.com/cape.png",
		}, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusOK, result.StatusCode)
		t.Equal("application/json", result.Header.Get("Content-Type"))
		body, _ := io.ReadAll(result.Body)
		t.JSONEq(`{
			"id": "mock-uuid",
			"name": "mock_username",
			"properties": [
				{
					"name": "textures",
					"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6Im1vY2stdXVpZCIsInByb2ZpbGVOYW1lIjoibW9ja191c2VybmFtZSIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9za2luLnBuZyIsIm1ldGFkYXRhIjp7Im1vZGVsIjoic2xpbSJ9fSwiQ0FQRSI6eyJ1cmwiOiJodHRwczovL2V4YW1wbGUuY29tL2NhcGUucG5nIn19fQ=="
				},
				{
					"name": "texturesParamName",
					"value": "texturesParamValue"
				}
			]
		}`, string(body))
	})

	t.Run("exists signed profile with skin", func() {
		req := httptest.NewRequest("GET", "http://chrly/profile/mock_username?unsigned=false", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{
			Uuid:      "mock-uuid",
			Username:  "mock_username",
			SkinUrl:   "https://example.com/skin.png",
			SkinModel: "slim",
		}, nil)
		t.TexturesSigner.On("SignTextures", "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6Im1vY2stdXVpZCIsInByb2ZpbGVOYW1lIjoibW9ja191c2VybmFtZSIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9za2luLnBuZyIsIm1ldGFkYXRhIjp7Im1vZGVsIjoic2xpbSJ9fX19").Return("mock signature", nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusOK, result.StatusCode)
		t.Equal("application/json", result.Header.Get("Content-Type"))
		body, _ := io.ReadAll(result.Body)
		t.JSONEq(`{
			"id": "mock-uuid",
			"name": "mock_username",
			"properties": [
				{
					"name": "textures",
                    "signature": "mock signature",
					"value": "eyJ0aW1lc3RhbXAiOjE2MTQyMTQyMjMwMDAsInByb2ZpbGVJZCI6Im1vY2stdXVpZCIsInByb2ZpbGVOYW1lIjoibW9ja191c2VybmFtZSIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9za2luLnBuZyIsIm1ldGFkYXRhIjp7Im1vZGVsIjoic2xpbSJ9fX19"
				},
				{
					"name": "texturesParamName",
					"value": "texturesParamValue"
				}
			]
		}`, string(body))
	})

	t.Run("not exists profile", func() {
		req := httptest.NewRequest("GET", "http://chrly/profile/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(nil, nil)

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusNotFound, result.StatusCode)
		body, _ := io.ReadAll(result.Body)
		t.Empty(body)
	})

	t.Run("err from profiles provider", func() {
		req := httptest.NewRequest("GET", "http://chrly/profile/mock_username", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(nil, errors.New("mock error"))

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusInternalServerError, result.StatusCode)
	})

	t.Run("err from textures signer", func() {
		req := httptest.NewRequest("GET", "http://chrly/profile/mock_username?unsigned=false", nil)
		w := httptest.NewRecorder()

		t.ProfilesProvider.On("FindProfileByUsername", mock.Anything, "mock_username", true).Return(&db.Profile{}, nil)
		t.TexturesSigner.On("SignTextures", mock.Anything).Return("", errors.New("mock error"))

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusInternalServerError, result.StatusCode)
	})
}

type signingKeyTestCase struct {
	Name       string
	KeyFormat  string
	BeforeTest func(suite *SkinsystemTestSuite)
	PanicErr   string
	AfterTest  func(suite *SkinsystemTestSuite, response *http.Response)
}

var signingKeyTestsCases = []*signingKeyTestCase{
	{
		Name:      "Get public key in DER format",
		KeyFormat: "DER",
		BeforeTest: func(suite *SkinsystemTestSuite) {
			pubPem, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANbUpVCZkMKpfvYZ08W3lumdAaYxLBnm\nUDlzHBQH3DpYef5WCO32TDU6feIJ58A0lAywgtZ4wwi2dGHOz/1hAvcCAwEAAQ==\n-----END PUBLIC KEY-----"))
			publicKey, _ := x509.ParsePKIXPublicKey(pubPem.Bytes)

			suite.TexturesSigner.On("GetPublicKey").Return(publicKey, nil)
		},
		AfterTest: func(suite *SkinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("application/octet-stream", response.Header.Get("Content-Type"))
			suite.Equal("attachment; filename=\"yggdrasil_session_pubkey.der\"", response.Header.Get("Content-Disposition"))
			body, _ := io.ReadAll(response.Body)
			suite.Equal([]byte{48, 92, 48, 13, 6, 9, 42, 134, 72, 134, 247, 13, 1, 1, 1, 5, 0, 3, 75, 0, 48, 72, 2, 65, 0, 214, 212, 165, 80, 153, 144, 194, 169, 126, 246, 25, 211, 197, 183, 150, 233, 157, 1, 166, 49, 44, 25, 230, 80, 57, 115, 28, 20, 7, 220, 58, 88, 121, 254, 86, 8, 237, 246, 76, 53, 58, 125, 226, 9, 231, 192, 52, 148, 12, 176, 130, 214, 120, 195, 8, 182, 116, 97, 206, 207, 253, 97, 2, 247, 2, 3, 1, 0, 1}, body)
		},
	},
	{
		Name:      "Get public key in PEM format",
		KeyFormat: "PEM",
		BeforeTest: func(suite *SkinsystemTestSuite) {
			pubPem, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANbUpVCZkMKpfvYZ08W3lumdAaYxLBnm\nUDlzHBQH3DpYef5WCO32TDU6feIJ58A0lAywgtZ4wwi2dGHOz/1hAvcCAwEAAQ==\n-----END PUBLIC KEY-----"))
			publicKey, _ := x509.ParsePKIXPublicKey(pubPem.Bytes)

			suite.TexturesSigner.On("GetPublicKey").Return(publicKey, nil)
		},
		AfterTest: func(suite *SkinsystemTestSuite, response *http.Response) {
			suite.Equal(200, response.StatusCode)
			suite.Equal("text/plain; charset=utf-8", response.Header.Get("Content-Type"))
			suite.Equal("attachment; filename=\"yggdrasil_session_pubkey.pem\"", response.Header.Get("Content-Disposition"))
			body, _ := io.ReadAll(response.Body)
			suite.Equal("-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANbUpVCZkMKpfvYZ08W3lumdAaYxLBnm\nUDlzHBQH3DpYef5WCO32TDU6feIJ58A0lAywgtZ4wwi2dGHOz/1hAvcCAwEAAQ==\n-----END PUBLIC KEY-----\n", string(body))
		},
	},
	{
		Name:      "Error while obtaining public key",
		KeyFormat: "DER",
		BeforeTest: func(suite *SkinsystemTestSuite) {
			suite.TexturesSigner.On("GetPublicKey").Return(nil, errors.New("textures signer error"))
		},
		PanicErr: "textures signer error",
	},
}

func (t *SkinsystemTestSuite) TestSignatureVerificationKey() {
	for _, testCase := range signingKeyTestsCases {
		t.Run(testCase.Name, func() {
			testCase.BeforeTest(t)

			req := httptest.NewRequest("GET", "http://chrly/signature-verification-key."+strings.ToLower(testCase.KeyFormat), nil)
			w := httptest.NewRecorder()

			if testCase.PanicErr != "" {
				t.PanicsWithError(testCase.PanicErr, func() {
					t.App.Handler().ServeHTTP(w, req)
				})
			} else {
				t.App.Handler().ServeHTTP(w, req)
				testCase.AfterTest(t, w.Result())
			}
		})
	}
}

func TestSkinsystem(t *testing.T) {
	suite.Run(t, new(SkinsystemTestSuite))
}

func TestParseUsername(t *testing.T) {
	assert := testify.New(t)
	assert.Equal("test", parseUsername("test.png"), "Function should trim .png at end")
	assert.Equal("test", parseUsername("test"), "Function should return string itself, if it not contains .png at end")
}
