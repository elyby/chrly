package signer

import (
	"context"
	"io"
	"strings"
)

type Signer interface {
	Sign(data io.Reader) ([]byte, error)
	GetPublicKey(format string) ([]byte, error)
}

type LocalSigner struct {
	Signer
}

func (s *LocalSigner) Sign(ctx context.Context, data string) (string, error) {
	signed, err := s.Signer.Sign(strings.NewReader(data))
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

func (s *LocalSigner) GetPublicKey(ctx context.Context, format string) (string, error) {
	publicKey, err := s.Signer.GetPublicKey(format)
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}
