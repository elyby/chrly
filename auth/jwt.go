package auth

import (
	"encoding/base64"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/mitchellh/go-homedir"
)

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
	if err := ioutil.WriteFile(getKeyPath(), key, 0600); err != nil {
		return err
	}

	return nil
}

func (t *JwtAuth) getSigningKey() ([]byte, error) {
	if t.signingKey != nil {
		return t.signingKey, nil
	}

	path := getKeyPath()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, &SigningKeyNotAvailable{}
		}

		return nil, err
	}

	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func createAppHomeDir() error {
	path := getAppHomeDirPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755) // rwx r-x r-x
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
	randLen := int(math.Ceil(float64(n) / 1.37)) // base64 will increase length in 1.37 times
	randBytes := make([]byte, randLen)
	rand.Read(randBytes)
	resBytes := make([]byte, n)
	base64.URLEncoding.Encode(resBytes, randBytes)

	return resBytes
}

type SigningKeyNotAvailable struct {
}

func (*SigningKeyNotAvailable) Error() string {
	return "Signing key not available"
}
