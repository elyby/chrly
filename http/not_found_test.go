package http

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	testify "github.com/stretchr/testify/assert"
)

func TestConfig_NotFound(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("GET", "http://skinsystem.ely.by/", nil)
	w := httptest.NewRecorder()

	(&Config{}).CreateHandler().ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(404, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))
	response, _ := ioutil.ReadAll(resp.Body)
	assert.JSONEq(`{
		"status": "404",
		"message": "Not Found"
	}`, string(response))
}
