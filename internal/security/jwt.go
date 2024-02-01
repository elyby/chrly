package security

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ely.by/chrly/internal/version"
)

var now = time.Now
var signingMethod = jwt.SigningMethodHS256

const scopesClaim = "scopes"

type Scope string

const (
	ProfileScope Scope = "profiles"
)

func NewJwt(key []byte) *Jwt {
	return &Jwt{
		Key: key,
	}
}

type Jwt struct {
	Key []byte
}

func (t *Jwt) NewToken(scopes ...Scope) (string, error) {
	if len(scopes) == 0 {
		return "", errors.New("you must specify at least one scope")
	}

	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims{
		"iss":       "chrly",
		"iat":       now().Unix(),
		scopesClaim: scopes,
	})
	token.Header["v"] = version.MajorVersion

	return token.SignedString(t.Key)
}

// Keep those names generic in order to reuse them in future for alternative authentication methods
var MissingAuthenticationError = errors.New("authentication value not provided")
var InvalidTokenError = errors.New("passed authentication value is invalid")

func (t *Jwt) Authenticate(req *http.Request) error {
	bearerToken := req.Header.Get("Authorization")
	if bearerToken == "" {
		return MissingAuthenticationError
	}

	if !strings.HasPrefix(strings.ToLower(bearerToken), "bearer ") {
		return InvalidTokenError
	}

	tokenStr := bearerToken[7:]
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return t.Key, nil
	})
	if err != nil {
		return errors.Join(InvalidTokenError, err)
	}

	if _, vHeaderExists := token.Header["v"]; !vHeaderExists {
		return errors.Join(InvalidTokenError, errors.New("missing v header"))
	}

	return nil
}
