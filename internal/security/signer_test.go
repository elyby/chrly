package security

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
)

type ConstantReader struct {
}

func (c *ConstantReader) Read(p []byte) (int, error) {
	return 1, nil
}

func TestSigner_SignTextures(t *testing.T) {
	randomReader = &ConstantReader{}

	rawKey, _ := pem.Decode([]byte("-----BEGIN RSA PRIVATE KEY-----\nMIIBOwIBAAJBANbUpVCZkMKpfvYZ08W3lumdAaYxLBnmUDlzHBQH3DpYef5WCO32\nTDU6feIJ58A0lAywgtZ4wwi2dGHOz/1hAvcCAwEAAQJAItaxSHTe6PKbyEU/9pxj\nONdhYRYwVLLo56gnMYhkyoEqaaMsfov8hhoepkYZBMvZFB2bDOsQ2SaJ+E2eiBO4\nAQIhAPssS0+BR9w0bOdmjGqmdE9NrN5UJQcOW13s29+6QzUBAiEA2vWOepA5Apiu\npEA3pwoGdkVCrNSnnKjDQzDXBnpd3/cCIEFNd9sY4qUG4FWdXN6RnmXL7Sj0uZfH\nDMwzu8rEM5sBAiEAhvdoDNqLmbMdq3c+FsPSOeL1d21Zp/JK8kbPtFmHNf8CIQDV\n6FSZDwvWfuxaM7BsycQONkjDBTPNu+lqctJBGnBv3A==\n-----END RSA PRIVATE KEY-----\n"))
	key, _ := x509.ParsePKCS1PrivateKey(rawKey.Bytes)

	signer := NewSigner(key)

	signature, err := signer.SignTextures(context.Background(), "eyJ0aW1lc3RhbXAiOjE2MTQzMDcxMzQsInByb2ZpbGVJZCI6ImZmYzhmZGM5NTgyNDUwOWU4YTU3Yzk5Yjk0MGZiOTk2IiwicHJvZmlsZU5hbWUiOiJFcmlja1NrcmF1Y2giLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly9lbHkuYnkvc3RvcmFnZS9za2lucy82OWM2NzQwZDI5OTNlNWQ2ZjZhN2ZjOTI0MjBlZmMyOS5wbmcifX0sImVseSI6dHJ1ZX0")
	require.NoError(t, err)
	require.Equal(t, "IyHCxTP5ITquEXTHcwCtLd08jWWy16JwlQeWg8naxhoAVQecHGRdzHRscuxtdq/446kmeox7h4EfRN2A2ZLL+A==", signature)
}

func TestSigner_GetPublicKey(t *testing.T) {
	randomReader = &ConstantReader{}

	rawKey, _ := pem.Decode([]byte("-----BEGIN RSA PRIVATE KEY-----\nMIIBOwIBAAJBANbUpVCZkMKpfvYZ08W3lumdAaYxLBnmUDlzHBQH3DpYef5WCO32\nTDU6feIJ58A0lAywgtZ4wwi2dGHOz/1hAvcCAwEAAQJAItaxSHTe6PKbyEU/9pxj\nONdhYRYwVLLo56gnMYhkyoEqaaMsfov8hhoepkYZBMvZFB2bDOsQ2SaJ+E2eiBO4\nAQIhAPssS0+BR9w0bOdmjGqmdE9NrN5UJQcOW13s29+6QzUBAiEA2vWOepA5Apiu\npEA3pwoGdkVCrNSnnKjDQzDXBnpd3/cCIEFNd9sY4qUG4FWdXN6RnmXL7Sj0uZfH\nDMwzu8rEM5sBAiEAhvdoDNqLmbMdq3c+FsPSOeL1d21Zp/JK8kbPtFmHNf8CIQDV\n6FSZDwvWfuxaM7BsycQONkjDBTPNu+lqctJBGnBv3A==\n-----END RSA PRIVATE KEY-----\n"))
	key, _ := x509.ParsePKCS1PrivateKey(rawKey.Bytes)

	signer := NewSigner(key)

	publicKey, err := signer.GetPublicKey(context.Background())
	require.NoError(t, err)
	require.IsType(t, &rsa.PublicKey{}, publicKey)
}
