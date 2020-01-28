package http

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type emitterMock struct {
	mock.Mock
}

func (e *emitterMock) Emit(name string, args ...interface{}) {
	e.Called((append([]interface{}{name}, args...))...)
}

func TestConfig_NotFound(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	NotFound(w, req)

	resp := w.Result()
	assert.Equal(404, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))
	response, _ := ioutil.ReadAll(resp.Body)
	assert.JSONEq(`{
		"status": "404",
		"message": "Not Found"
	}`, string(response))
}
