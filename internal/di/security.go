package di

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log/slog"

	"ely.by/chrly/internal/client/signer"
	"ely.by/chrly/internal/http"
	"ely.by/chrly/internal/security"

	"github.com/defval/di"
	"github.com/spf13/viper"
)

var securityDiOptions = di.Options(
	di.Provide(newSigner,
		di.As(new(http.Signer)),
		di.As(new(signer.Signer)),
	),
	di.Provide(newSignerService),
)

func newSigner(config *viper.Viper) (*security.Signer, error) {
	var privateKey *rsa.PrivateKey
	var err error

	keyStr := config.GetString("chrly.signing.key")
	if keyStr == "" {
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}

		slog.Warn("A private signing key has been generated. To make it permanent, specify the valid RSA private key in the config parameter chrly.signing.key")
	} else {
		keyBytes := []byte(keyStr)
		rawPem, _ := pem.Decode(keyBytes)
		if rawPem == nil {
			return nil, errors.New("unable to decode pem key")
		}

		privateKey, err = x509.ParsePKCS1PrivateKey(rawPem.Bytes)
		if err != nil {
			return nil, err
		}
	}

	return security.NewSigner(privateKey), nil
}

func newSignerService(s signer.Signer) http.SignerService {
	return &signer.LocalSigner{
		Signer: s,
	}
}
