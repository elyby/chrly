package auth

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/afero"

	testify "github.com/stretchr/testify/assert"
)

const jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxNTE2NjU4MTkzIiwic2NvcGVzIjoic2tpbiJ9.agbBS0qdyYMBaVfTZJAZcTTRgW1Y0kZty4H3N2JHBO8"

func TestJwtAuth_NewToken_Success(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	fs.Mkdir(getAppHomeDirPath(), 0755)
	afero.WriteFile(fs, getKeyPath(), []byte("secret"), 0600)

	jwt := &JwtAuth{}
	token, err := jwt.NewToken(SkinScope)
	assert.Nil(err)
	assert.NotNil(token)
}

func TestJwtAuth_NewToken_KeyNotAvailable(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	fs = afero.NewMemMapFs()

	jwt := &JwtAuth{}
	token, err := jwt.NewToken(SkinScope)
	assert.IsType(&SigningKeyNotAvailable{}, err)
	assert.Nil(token)
}

func TestJwtAuth_GenerateSigningKey_KeyNotExists(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	jwt := &JwtAuth{}
	err := jwt.GenerateSigningKey()
	assert.Nil(err)
	if _, err := fs.Stat(getAppHomeDirPath()); err != nil {
		assert.Fail("directory not created")
	}

	if _, err := fs.Stat(getKeyPath()); err != nil {
		assert.Fail("signing file not created")
	}

	content, _ := afero.ReadFile(fs, getKeyPath())
	assert.Len(content, 64)
}

func TestJwtAuth_GenerateSigningKey_KeyExists(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	fs.Mkdir(getAppHomeDirPath(), 0755)
	afero.WriteFile(fs, getKeyPath(), []byte("secret"), 0600)

	jwt := &JwtAuth{}
	err := jwt.GenerateSigningKey()
	assert.Nil(err)
	if _, err := fs.Stat(getAppHomeDirPath()); err != nil {
		assert.Fail("directory not created")
	}

	if _, err := fs.Stat(getKeyPath()); err != nil {
		assert.Fail("signing file not created")
	}

	content, _ := afero.ReadFile(fs, getKeyPath())
	assert.NotEqual([]byte("secret"), content)
}

func TestJwtAuth_Check_EmptyRequest(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	jwt := &JwtAuth{}

	err := jwt.Check(req)
	assert.IsType(&Unauthorized{}, err)
	assert.EqualError(err, "Authentication header not presented")
}

func TestJwtAuth_Check_NonBearer(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "this is not jwt")
	jwt := &JwtAuth{}

	err := jwt.Check(req)
	assert.IsType(&Unauthorized{}, err)
	assert.EqualError(err, "Cannot recognize JWT token in passed value")
}

func TestJwtAuth_Check_BearerButNotJwt(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "Bearer thisIs.Not.Jwt")
	jwt := &JwtAuth{}

	err := jwt.Check(req)
	assert.IsType(&Unauthorized{}, err)
	assert.EqualError(err, "Cannot parse passed JWT token")
}

func TestJwtAuth_Check_SecretNotAvailable(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "Bearer " + jwt)
	jwt := &JwtAuth{}

	err := jwt.Check(req)
	assert.IsType(&SigningKeyNotAvailable{}, err)
}

func TestJwtAuth_Check_SecretInvalid(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "Bearer " + jwt)
	jwt := &JwtAuth{[]byte("this is another secret")}

	err := jwt.Check(req)
	assert.IsType(&Unauthorized{}, err)
	assert.EqualError(err, "JWT token have invalid signature. It corrupted or expired.")
}

func TestJwtAuth_Check_Valid(t *testing.T) {
	clearFs()
	assert := testify.New(t)

	req := httptest.NewRequest("POST", "http://localhost", nil)
	req.Header.Add("Authorization", "Bearer " + jwt)
	jwt := &JwtAuth{[]byte("secret")}

	err := jwt.Check(req)
	assert.Nil(err)
}

func TestJwtAuth_generateRandomBytes(t *testing.T) {
	assert := testify.New(t)
	lengthMap := []int{12, 20, 24, 30, 32, 48, 50, 64}
	for _, length := range lengthMap {
		bytes := generateRandomBytes(length)
		assert.Len(bytes, length)
		assert.False(strings.HasSuffix(string(bytes), "="), "secret key should not ends with '=' character")
	}
}

func clearFs() {
	fs = afero.NewMemMapFs()
}
