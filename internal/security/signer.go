package security

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
)

var randomReader = rand.Reader
var invalidKeyFormat = errors.New(`invalid key format: it should be"der" or "pem"`)

func NewSigner(key *rsa.PrivateKey) *Signer {
	return &Signer{Key: key}
}

type Signer struct {
	Key *rsa.PrivateKey
}

func (s *Signer) Sign(data io.Reader) ([]byte, error) {
	messageHash := sha1.New()
	_, err := io.Copy(messageHash, data)
	if err != nil {
		return nil, err
	}

	messageHashSum := messageHash.Sum(nil)
	signature, err := rsa.SignPKCS1v15(randomReader, s.Key, crypto.SHA1, messageHashSum)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (s *Signer) GetPublicKey(format string) ([]byte, error) {
	if format != "der" && format != "pem" {
		return nil, invalidKeyFormat
	}

	asn1Bytes, err := x509.MarshalPKIXPublicKey(s.Key.Public())
	if err != nil {
		return nil, err
	}

	if format == "pem" {
		return pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: asn1Bytes,
		}), nil
	}

	return asn1Bytes, nil
}
