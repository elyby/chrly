package security

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const jwtString = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsInYiOjV9.eyJpYXQiOjE3MDY3ODY3NzUsImlzcyI6ImNocmx5Iiwic2NvcGVzIjpbInByb2ZpbGVzIl19.LrXrKo5iRFFHCDlMsVDhmJJheZqxbxuEVXB4XswHFKY"

func TestJwtAuth_NewToken(t *testing.T) {
	jwt := NewJwt([]byte("secret"))
	now = func() time.Time {
		return time.Date(2024, 2, 1, 11, 26, 15, 0, time.UTC)
	}

	t.Run("with scope", func(t *testing.T) {
		token, err := jwt.NewToken(ProfileScope, "custom-scope")
		require.NoError(t, err)
		require.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsInYiOjV9.eyJpYXQiOjE3MDY3ODY3NzUsImlzcyI6ImNocmx5Iiwic2NvcGVzIjpbInByb2ZpbGVzIiwiY3VzdG9tLXNjb3BlIl19.Iq673YyWWkJZjIkBmKYRN8Lx9qoD39S_e-MegG0aORM", token)
	})

	t.Run("no scopes", func(t *testing.T) {
		token, err := jwt.NewToken()
		require.Error(t, err)
		require.Empty(t, token)
	})
}

func TestJwtAuth_Authenticate(t *testing.T) {
	jwt := NewJwt([]byte("secret"))
	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "Bearer "+jwtString)
		err := jwt.Authenticate(req)
		require.NoError(t, err)
	})

	t.Run("request without auth header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://localhost", nil)
		err := jwt.Authenticate(req)
		require.ErrorIs(t, err, MissingAuthenticationError)
	})

	t.Run("no bearer token prefix", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "trash")
		err := jwt.Authenticate(req)
		require.ErrorIs(t, err, InvalidTokenError)
	})

	t.Run("bearer token but not jwt", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "Bearer seems.like.jwt")
		err := jwt.Authenticate(req)
		require.ErrorIs(t, err, InvalidTokenError)
	})

	t.Run("invalid signature", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "Bearer "+jwtString+"123")
		err := jwt.Authenticate(req)
		require.ErrorIs(t, err, InvalidTokenError)
	})

	t.Run("missing v header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE3MDY3ODY3NzUsImlzcyI6ImNocmx5Iiwic2NvcGVzIjpbInByb2ZpbGVzIl19.zOX2ZKyU37kjwt1p9uCHxALxWQD2UC0wWcAcNvBXGq0")
		err := jwt.Authenticate(req)
		require.ErrorIs(t, err, InvalidTokenError)
		require.ErrorContains(t, err, "missing v header")
	})
}
