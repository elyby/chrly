package http

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateRequestEventsMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	isHandlerCalled := false
	middlewareFunc := CreateRequestEventsMiddleware()
	middlewareFunc.Middleware(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(400)
		isHandlerCalled = true
	})).ServeHTTP(resp, req)

	if !isHandlerCalled {
		t.Fatal("Handler isn't called from the middleware")
	}
}

type authCheckerMock struct {
	mock.Mock
}

func (m *authCheckerMock) Authenticate(req *http.Request) error {
	args := m.Called(req)
	return args.Error(0)
}

func TestCreateAuthenticationMiddleware(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		resp := httptest.NewRecorder()

		auth := &authCheckerMock{}
		auth.On("Authenticate", req).Once().Return(nil)

		isHandlerCalled := false
		middlewareFunc := CreateAuthenticationMiddleware(auth)
		middlewareFunc.Middleware(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			isHandlerCalled = true
		})).ServeHTTP(resp, req)

		testify.True(t, isHandlerCalled, "Handler isn't called from the middleware")

		auth.AssertExpectations(t)
	})

	t.Run("fail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		resp := httptest.NewRecorder()

		auth := &authCheckerMock{}
		auth.On("Authenticate", req).Once().Return(errors.New("error reason"))

		isHandlerCalled := false
		middlewareFunc := CreateAuthenticationMiddleware(auth)
		middlewareFunc.Middleware(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			isHandlerCalled = true
		})).ServeHTTP(resp, req)

		testify.False(t, isHandlerCalled, "Handler shouldn't be called")
		testify.Equal(t, 403, resp.Code)
		body, _ := ioutil.ReadAll(resp.Body)
		testify.JSONEq(t, `{
			"error": "error reason"
		}`, string(body))

		auth.AssertExpectations(t)
	})
}

func TestNotFoundHandler(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	NotFoundHandler(w, req)

	resp := w.Result()
	assert.Equal(404, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))
	response, _ := ioutil.ReadAll(resp.Body)
	assert.JSONEq(`{
		"status": "404",
		"message": "Not Found"
	}`, string(response))
}
