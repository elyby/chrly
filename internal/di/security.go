package di

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"

	signerClient "ely.by/chrly/internal/client/signer"
	"ely.by/chrly/internal/http"
	"ely.by/chrly/internal/security"

	"github.com/defval/di"
	"github.com/spf13/viper"
)

var securityDiOptions = di.Options(
	di.Provide(newSigner,
		di.As(new(http.Signer)),
		di.As(new(signerClient.Signer)),
	),
	di.Provide(newSignerService),
)

func newSigner(config *viper.Viper) (*security.Signer, error) {
	keyStr := config.GetString("chrly.signing.key")
	if keyStr == "" {
		// TODO: log a message about the generated signing key and the way to specify it permanently
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}

		return security.NewSigner(privateKey), nil
	}

	var keyBytes []byte
	if strings.HasPrefix(keyStr, "base64:") {
		base64Value := keyStr[7:]
		decodedKey, err := base64.URLEncoding.DecodeString(base64Value)
		if err != nil {
			return nil, err
		}

		keyBytes = decodedKey
	} else {
		keyBytes = []byte(keyStr)
	}

	rawPem, _ := pem.Decode(keyBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(rawPem.Bytes)
	if err != nil {
		return nil, err
	}

	return security.NewSigner(privateKey), nil
}

func newSignerService(signer signerClient.Signer) http.SignerService {
	return &signerClient.LocalSigner{
		Signer: signer,
	}
}
