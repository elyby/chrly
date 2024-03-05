package http

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SignerMock struct {
	mock.Mock
}

func (m *SignerMock) Sign(data io.Reader) ([]byte, error) {
	args := m.Called(data)
	var result []byte
	if casted, ok := args.Get(0).([]byte); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *SignerMock) GetPublicKey(format string) ([]byte, error) {
	args := m.Called(format)
	var result []byte
	if casted, ok := args.Get(0).([]byte); ok {
		result = casted
	}

	return result, args.Error(1)
}

type SignerApiTestSuite struct {
	suite.Suite

	App *SignerApi

	Signer *SignerMock
}

func (t *SignerApiTestSuite) SetupSubTest() {
	t.Signer = &SignerMock{}

	t.App = &SignerApi{
		Signer: t.Signer,
	}
}

func (t *SignerApiTestSuite) TearDownSubTest() {
	t.Signer.AssertExpectations(t.T())
}

func (t *SignerApiTestSuite) TestSign() {
	t.Run("successfully sign", func() {
		signature := []byte("mock signature")
		t.Signer.On("Sign", mock.Anything).Return(signature, nil).Run(func(args mock.Arguments) {
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, args.Get(0).(io.Reader))
			r, _ := io.ReadAll(buf)

			t.Equal([]byte("mock body to sign"), r)
		})

		req := httptest.NewRequest("POST", "http://chrly/", strings.NewReader("mock body to sign"))
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusOK, result.StatusCode)
		t.Equal("application/octet-stream+base64", result.Header.Get("Content-Type"))
		body, _ := io.ReadAll(result.Body)
		t.Equal([]byte{0x62, 0x57, 0x39, 0x6a, 0x61, 0x79, 0x42, 0x7a, 0x61, 0x57, 0x64, 0x75, 0x59, 0x58, 0x52, 0x31, 0x63, 0x6d, 0x55, 0x3d}, body)
	})

	t.Run("handle error during sign", func() {
		t.Signer.On("Sign", mock.Anything).Return(nil, errors.New("mock error"))

		req := httptest.NewRequest("POST", "http://chrly/", strings.NewReader("mock body to sign"))
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusInternalServerError, result.StatusCode)
	})
}

func (t *SignerApiTestSuite) TestGetPublicKey() {
	t.Run("in pem format", func() {
		publicKey := []byte("mock public key in pem format")
		t.Signer.On("GetPublicKey", "pem").Return(publicKey, nil)

		req := httptest.NewRequest("GET", "http://chrly/public-key.pem", nil)
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusOK, result.StatusCode)
		t.Equal("application/x-pem-file", result.Header.Get("Content-Type"))
		t.Equal(`attachment; filename="yggdrasil_session_pubkey.pem"`, result.Header.Get("Content-Disposition"))
		body, _ := io.ReadAll(result.Body)
		t.Equal(publicKey, body)
	})

	t.Run("in der format", func() {
		publicKey := []byte("mock public key in der format")
		t.Signer.On("GetPublicKey", "der").Return(publicKey, nil)

		req := httptest.NewRequest("GET", "http://chrly/public-key.der", nil)
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusOK, result.StatusCode)
		t.Equal("application/octet-stream", result.Header.Get("Content-Type"))
		t.Equal(`attachment; filename="yggdrasil_session_pubkey.der"`, result.Header.Get("Content-Disposition"))
		body, _ := io.ReadAll(result.Body)
		t.Equal(publicKey, body)
	})

	t.Run("handle error", func() {
		t.Signer.On("GetPublicKey", "pem").Return(nil, errors.New("mock error"))

		req := httptest.NewRequest("GET", "http://chrly/public-key.pem", nil)
		w := httptest.NewRecorder()

		t.App.Handler().ServeHTTP(w, req)

		result := w.Result()
		t.Equal(http.StatusInternalServerError, result.StatusCode)
	})
}

func TestSignerApi(t *testing.T) {
	suite.Run(t, new(SignerApiTestSuite))
}
