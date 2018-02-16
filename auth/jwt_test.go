package auth

import (
	"net/http/httptest"
	"testing"

	testify "github.com/stretchr/testify/assert"
)

const jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxNTE2NjU4MTkzIiwic2NvcGVzIjoic2tpbiJ9.agbBS0qdyYMBaVfTZJAZcTTRgW1Y0kZty4H3N2JHBO8"

func TestJwtAuth_NewToken_Success(t *testing.T) {
	assert := testify.New(t)

	jwt := &JwtAuth{[]byte("secret")}
	token, err := jwt.NewToken(SkinScope)
	assert.Nil(err)
	assert.NotNil(token)
}

func TestJwtAuth_NewToken_KeyNotAvailable(t *testing.T) {
	assert := testify.New(t)

	jwt := &JwtAuth{}
	token, err := jwt.NewToken(SkinScope)
	assert.Error(err, "signing key not available")
	assert.Nil(token)
}

func TestJwtAuth_Check_EmptyRequest(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	jwt := &JwtAuth{[]byte("secret")}

	err := jwt.Check(req)
	assert.IsType(&Unauthorized{}, err)
	assert.EqualError(err, "Authentication header not presented")
}

func TestJwtAuth_Check_NonBearer(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "this is not jwt")
	jwt := &JwtAuth{[]byte("secret")}

	err := jwt.Check(req)
	assert.IsType(&Unauthorized{}, err)
	assert.EqualError(err, "Cannot recognize JWT token in passed value")
}

func TestJwtAuth_Check_BearerButNotJwt(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "Bearer thisIs.Not.Jwt")
	jwt := &JwtAuth{[]byte("secret")}

	err := jwt.Check(req)
	assert.IsType(&Unauthorized{}, err)
	assert.EqualError(err, "Cannot parse passed JWT token")
}

func TestJwtAuth_Check_SecretNotAvailable(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "Bearer " + jwt)
	jwt := &JwtAuth{}

	err := jwt.Check(req)
	assert.Error(err, "Signing key not set")
}

func TestJwtAuth_Check_SecretInvalid(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "Bearer " + jwt)
	jwt := &JwtAuth{[]byte("this is another secret")}

	err := jwt.Check(req)
	assert.IsType(&Unauthorized{}, err)
	assert.EqualError(err, "JWT token have invalid signature. It may be corrupted or expired.")
}

func TestJwtAuth_Check_Valid(t *testing.T) {
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "Bearer " + jwt)
	jwt := &JwtAuth{[]byte("secret")}

	err := jwt.Check(req)
	assert.Nil(err)
}
