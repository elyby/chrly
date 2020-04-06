package http

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxNTE2NjU4MTkzIiwic2NvcGVzIjoic2tpbiJ9.agbBS0qdyYMBaVfTZJAZcTTRgW1Y0kZty4H3N2JHBO8"

func TestJwtAuth_NewToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		jwt := &JwtAuth{Key: []byte("secret")}
		token, err := jwt.NewToken(SkinScope)
		assert.Nil(t, err)
		assert.NotNil(t, token)
	})

	t.Run("key not provided", func(t *testing.T) {
		jwt := &JwtAuth{}
		token, err := jwt.NewToken(SkinScope)
		assert.Error(t, err, "signing key not available")
		assert.Nil(t, token)
	})
}

func TestJwtAuth_Authenticate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emitter := &emitterMock{}
		emitter.On("Emit", "authentication:success")

		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "Bearer " + jwt)
		jwt := &JwtAuth{Key: []byte("secret"), Emitter: emitter}

		err := jwt.Authenticate(req)
		assert.Nil(t, err)

		emitter.AssertExpectations(t)
	})

	t.Run("request without auth header", func(t *testing.T) {
		emitter := &emitterMock{}
		emitter.On("Emit", "authentication:error", mock.MatchedBy(func(err error) bool {
			assert.Error(t, err, "Authentication header not presented")
			return true
		}))

		req := httptest.NewRequest("POST", "http://localhost", nil)
		jwt := &JwtAuth{Key: []byte("secret"), Emitter: emitter}

		err := jwt.Authenticate(req)
		assert.Error(t, err, "Authentication header not presented")

		emitter.AssertExpectations(t)
	})

	t.Run("no bearer token prefix", func(t *testing.T) {
		emitter := &emitterMock{}
		emitter.On("Emit", "authentication:error", mock.MatchedBy(func(err error) bool {
			assert.Error(t, err, "Cannot recognize JWT token in passed value")
			return true
		}))

		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "this is not jwt")
		jwt := &JwtAuth{Key: []byte("secret"), Emitter: emitter}

		err := jwt.Authenticate(req)
		assert.Error(t, err, "Cannot recognize JWT token in passed value")

		emitter.AssertExpectations(t)
	})

	t.Run("bearer token but not jwt", func(t *testing.T) {
		emitter := &emitterMock{}
		emitter.On("Emit", "authentication:error", mock.MatchedBy(func(err error) bool {
			assert.Error(t, err, "Cannot parse passed JWT token")
			return true
		}))

		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "Bearer thisIs.Not.Jwt")
		jwt := &JwtAuth{Key: []byte("secret"), Emitter: emitter}

		err := jwt.Authenticate(req)
		assert.Error(t, err, "Cannot parse passed JWT token")

		emitter.AssertExpectations(t)
	})

	t.Run("when secret is not set", func(t *testing.T) {
		emitter := &emitterMock{}
		emitter.On("Emit", "authentication:error", mock.MatchedBy(func(err error) bool {
			assert.Error(t, err, "Signing key not set")
			return true
		}))

		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "Bearer " + jwt)
		jwt := &JwtAuth{Emitter: emitter}

		err := jwt.Authenticate(req)
		assert.Error(t, err, "Signing key not set")

		emitter.AssertExpectations(t)
	})

	t.Run("invalid signature", func(t *testing.T) {
		emitter := &emitterMock{}
		emitter.On("Emit", "authentication:error", mock.MatchedBy(func(err error) bool {
			assert.Error(t, err, "JWT token have invalid signature. It may be corrupted or expired")
			return true
		}))

		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "Bearer " + jwt)
		jwt := &JwtAuth{Key: []byte("this is another secret"), Emitter: emitter}

		err := jwt.Authenticate(req)
		assert.Error(t, err, "JWT token have invalid signature. It may be corrupted or expired")

		emitter.AssertExpectations(t)
	})
}
