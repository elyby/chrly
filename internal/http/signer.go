package http

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"

	"ely.by/chrly/internal/otel"
)

type Signer interface {
	Sign(data io.Reader) ([]byte, error)
	GetPublicKey(format string) ([]byte, error)
}

func NewSignerApi(signer Signer) (*SignerApi, error) {
	metrics, err := newSignerApiMetrics(otel.GetMeter())
	if err != nil {
		return nil, err
	}

	return &SignerApi{
		Signer:  signer,
		metrics: metrics,
	}, nil
}

type SignerApi struct {
	Signer

	metrics *signerApiMetrics
}

func (s *SignerApi) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", s.signHandler).Methods(http.MethodPost)
	router.HandleFunc("/public-key.{format:(?:pem|der)}", s.getPublicKeyHandler).Methods(http.MethodGet)

	return router
}

func (s *SignerApi) signHandler(resp http.ResponseWriter, req *http.Request) {
	signature, err := s.Signer.Sign(req.Body)
	if err != nil {
		apiServerError(resp, req, fmt.Errorf("unable to sign the value: %w", err))
		return
	}

	resp.Header().Set("Content-Type", "application/octet-stream+base64")

	buf := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
	base64.StdEncoding.Encode(buf, signature)
	_, _ = resp.Write(buf)
}

func (s *SignerApi) getPublicKeyHandler(resp http.ResponseWriter, req *http.Request) {
	format := mux.Vars(req)["format"]
	publicKey, err := s.Signer.GetPublicKey(format)
	if err != nil {
		apiServerError(resp, req, fmt.Errorf("unable to retrieve public key: %w", err))
		return
	}

	if format == "pem" {
		resp.Header().Set("Content-Type", "application/x-pem-file")
		resp.Header().Set("Content-Disposition", `attachment; filename="yggdrasil_session_pubkey.pem"`)
	} else {
		resp.Header().Set("Content-Type", "application/octet-stream")
		resp.Header().Set("Content-Disposition", `attachment; filename="yggdrasil_session_pubkey.der"`)
	}

	_, _ = resp.Write(publicKey)
}

func newSignerApiMetrics(meter metric.Meter) (*signerApiMetrics, error) {
	m := &signerApiMetrics{}
	var errors, err error

	m.SignRequest, err = meter.Int64Counter("chrly.app.signer.sign.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	m.GetPublicKeyRequest, err = meter.Int64Counter("chrly.app.signer.get_public_key.request", metric.WithUnit("{request}"))
	errors = multierr.Append(errors, err)

	return m, errors
}

type signerApiMetrics struct {
	SignRequest         metric.Int64Counter
	GetPublicKeyRequest metric.Int64Counter
}
