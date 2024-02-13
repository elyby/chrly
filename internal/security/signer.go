package security

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
)

var randomReader = rand.Reader

func NewSigner(key *rsa.PrivateKey) *Signer {
	return &Signer{Key: key}
}

type Signer struct {
	Key *rsa.PrivateKey
}

func (s *Signer) SignTextures(ctx context.Context, textures string) (string, error) {
	message := []byte(textures)
	messageHash := sha1.New()
	_, err := messageHash.Write(message)
	if err != nil {
		return "", err
	}

	messageHashSum := messageHash.Sum(nil)
	signature, err := rsa.SignPKCS1v15(randomReader, s.Key, crypto.SHA1, messageHashSum)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *Signer) GetPublicKey(ctx context.Context) (*rsa.PublicKey, error) {
	return &s.Key.PublicKey, nil
}
