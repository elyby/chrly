package signer

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"errors"
)

var randomReader = rand.Reader

type Signer struct {
	Key *rsa.PrivateKey
}

func (s *Signer) SignTextures(textures string) (string, error) {
	if s.Key == nil {
		return "", errors.New("Key is empty")
	}

	message := []byte(textures)
	messageHash := sha1.New()
	_, _ = messageHash.Write(message)
	messageHashSum := messageHash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(randomReader, s.Key, crypto.SHA1, messageHashSum)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *Signer) GetPublicKey() (*rsa.PublicKey, error) {
	if s.Key == nil {
		return nil, errors.New("Key is empty")
	}

	return &s.Key.PublicKey, nil
}
