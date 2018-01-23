package auth

import (
	"encoding/base64"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

var hashAlg = crypto.SigningMethodHS256

const appHomeDirName = ".minecraft-skinsystem"
const scopesClaim = "scopes"

type Scope string

var (
	SkinScope = Scope("skin")
)

type JwtAuth struct {
	signingKey []byte
}

func (t *JwtAuth) NewToken(scopes ...Scope) ([]byte, error) {
	key, err := t.getSigningKey()
	if err != nil {
		return nil, err
	}

	claims := jws.Claims{}
	claims.Set(scopesClaim, scopes)
	claims.SetIssuedAt(time.Now())
	encoder := jws.NewJWT(claims, hashAlg)
	token, err := encoder.Serialize(key)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (t *JwtAuth) GenerateSigningKey() error {
	if err := createAppHomeDir(); err != nil {
		return err
	}

	key := generateRandomBytes(64)
	if err := afero.WriteFile(fs, getKeyPath(), key, 0600); err != nil {
		return err
	}

	return nil
}

func (t *JwtAuth) Check(req *http.Request) error {
	bearerToken := req.Header.Get("Authorization")
	if bearerToken == "" {
		return &Unauthorized{"Authentication header not presented"}
	}

	if !strings.EqualFold(bearerToken[0:7], "BEARER ") {
		return &Unauthorized{"Cannot recognize JWT token in passed value"}
	}

	tokenStr := bearerToken[7:]
	token, err := jws.ParseJWT([]byte(tokenStr))
	if err != nil {
		return &Unauthorized{"Cannot parse passed JWT token"}
	}

	signKey, err := t.getSigningKey()
	if err != nil {
		return err
	}

	err = token.Validate(signKey, hashAlg)
	if err != nil {
		return &Unauthorized{"JWT token have invalid signature. It corrupted or expired."}
	}

	return nil
}

func (t *JwtAuth) getSigningKey() ([]byte, error) {
	if t.signingKey == nil {
		path := getKeyPath()
		if _, err := fs.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return nil, &SigningKeyNotAvailable{}
			}

			return nil, err
		}

		key, err := afero.ReadFile(fs, path)
		if err != nil {
			return nil, err
		}

		t.signingKey = key
	}

	return t.signingKey, nil
}

func createAppHomeDir() error {
	path := getAppHomeDirPath()
	if _, err := fs.Stat(path); os.IsNotExist(err) {
		err := fs.Mkdir(path, 0755) // rwx r-x r-x
		if err != nil {
			return err
		}
	}

	return nil
}

func getAppHomeDirPath() string {
	path, err := homedir.Expand("~/" + appHomeDirName)
	if err != nil {
		panic(err)
	}

	return path
}

func getKeyPath() string {
	return getAppHomeDirPath() + "/jwt-key"
}

func generateRandomBytes(n int) []byte {
	// base64 will increase length in 1.37 times
	// +1 is needed to ensure, that after base64 we will do not have any '===' characters
	randLen := int(math.Ceil(float64(n) / 1.37)) + 1
	randBytes := make([]byte, randLen)
	rand.Read(randBytes)
	// +5 is needed to have additional buffer for the next set of XX=== characters
	resBytes := make([]byte, n + 5)
	base64.URLEncoding.Encode(resBytes, randBytes)

	return resBytes[:n]
}

type Unauthorized struct {
	Reason string
}

func (e *Unauthorized) Error() string {
	if e.Reason != "" {
		return e.Reason
	}

	return "Unauthorized"
}

type SigningKeyNotAvailable struct {
}

func (*SigningKeyNotAvailable) Error() string {
	return "Signing key not available"
}
