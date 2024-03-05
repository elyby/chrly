package http

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	testify "github.com/stretchr/testify/require"

	"ely.by/chrly/internal/security"
)

type authCheckerMock struct {
	mock.Mock
}

func (m *authCheckerMock) Authenticate(req *http.Request, scope security.Scope) error {
	return m.Called(req, scope).Error(0)
}

func TestAuthenticationMiddleware(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://example.com", nil)
		resp := httptest.NewRecorder()

		auth := &authCheckerMock{}
		auth.On("Authenticate", req, security.Scope("mock")).Once().Return(nil)

		isHandlerCalled := false
		middlewareFunc := NewAuthenticationMiddleware(auth, "mock")
		middlewareFunc.Middleware(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			isHandlerCalled = true
		})).ServeHTTP(resp, req)

		testify.True(t, isHandlerCalled, "Handler isn't called from the middleware")

		auth.AssertExpectations(t)
	})

	t.Run("fail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://example.com", nil)
		resp := httptest.NewRecorder()

		auth := &authCheckerMock{}
		auth.On("Authenticate", req, security.Scope("mock")).Once().Return(errors.New("error reason"))

		isHandlerCalled := false
		middlewareFunc := NewAuthenticationMiddleware(auth, "mock")
		middlewareFunc.Middleware(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			isHandlerCalled = true
		})).ServeHTTP(resp, req)

		testify.False(t, isHandlerCalled, "Handler shouldn't be called")
		testify.Equal(t, 403, resp.Code)
		body, _ := io.ReadAll(resp.Body)
		testify.JSONEq(t, `{
			"error": "error reason"
		}`, string(body))

		auth.AssertExpectations(t)
	})
}

func TestConditionalMiddleware(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://example.com", nil)
		resp := httptest.NewRecorder()

		isNestedMiddlewareCalled := false
		isHandlerCalled := false
		NewConditionalMiddleware(
			func(req *http.Request) bool {
				return true
			},
			func(handler http.Handler) http.Handler {
				isNestedMiddlewareCalled = true
				return handler
			},
		).Middleware(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			isHandlerCalled = true
		})).ServeHTTP(resp, req)

		testify.True(t, isNestedMiddlewareCalled, "Nested middleware wasn't called")
		testify.True(t, isHandlerCalled, "Handler wasn't called from the middleware")
	})

	t.Run("false", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://example.com", nil)
		resp := httptest.NewRecorder()

		isNestedMiddlewareCalled := false
		isHandlerCalled := false
		NewConditionalMiddleware(
			func(req *http.Request) bool {
				return false
			},
			func(handler http.Handler) http.Handler {
				isNestedMiddlewareCalled = true
				return handler
			},
		).Middleware(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			isHandlerCalled = true
		})).ServeHTTP(resp, req)

		testify.False(t, isNestedMiddlewareCalled, "Nested middleware shouldn't be called")
		testify.True(t, isHandlerCalled, "Handler wasn't called from the middleware")
	})
}

func TestNotFoundHandler(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("GET", "https://example.com", nil)
	w := httptest.NewRecorder()

	NotFoundHandler(w, req)

	resp := w.Result()
	assert.Equal(404, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))
	response, _ := io.ReadAll(resp.Body)
	assert.JSONEq(`{
		"status": "404",
		"message": "Not Found"
	}`, string(response))
}
