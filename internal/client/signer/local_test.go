package signer

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SignerMock struct {
	mock.Mock
}

func (m *SignerMock) Sign(data io.Reader) ([]byte, error) {
	args := m.Called(data)
	var result []byte
	if casted, ok := args.Get(0).([]byte); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *SignerMock) GetPublicKey(format string) ([]byte, error) {
	args := m.Called(format)
	var result []byte
	if casted, ok := args.Get(0).([]byte); ok {
		result = casted
	}

	return result, args.Error(1)
}

type LocalSignerServiceTestSuite struct {
	suite.Suite

	Service *LocalSigner

	Signer *SignerMock
}

func (t *LocalSignerServiceTestSuite) SetupSubTest() {
	t.Signer = &SignerMock{}

	t.Service = &LocalSigner{
		Signer: t.Signer,
	}
}

func (t *LocalSignerServiceTestSuite) TearDownSubTest() {
	t.Signer.AssertExpectations(t.T())
}

func (t *LocalSignerServiceTestSuite) TestSign() {
	t.Run("successfully sign", func() {
		signature := []byte("mock signature")
		t.Signer.On("Sign", mock.Anything).Return(signature, nil).Run(func(args mock.Arguments) {
			r, _ := io.ReadAll(args.Get(0).(io.Reader))
			t.Equal([]byte("mock body to sign"), r)
		})

		result, err := t.Service.Sign(context.Background(), "mock body to sign")
		t.NoError(err)
		t.Equal(string(signature), result)
	})

	t.Run("handle error during sign", func() {
		expectedErr := errors.New("mock error")
		t.Signer.On("Sign", mock.Anything).Return(nil, expectedErr)

		result, err := t.Service.Sign(context.Background(), "mock body to sign")
		t.Error(err)
		t.Same(expectedErr, err)
		t.Empty(result)
	})
}

func (t *LocalSignerServiceTestSuite) TestGetPublicKey() {
	t.Run("successfully get", func() {
		publicKey := []byte("mock public key")
		t.Signer.On("GetPublicKey", "pem").Return(publicKey, nil)

		result, err := t.Service.GetPublicKey(context.Background(), "pem")
		t.NoError(err)
		t.Equal(string(publicKey), result)
	})

	t.Run("handle error", func() {
		expectedErr := errors.New("mock error")
		t.Signer.On("GetPublicKey", "pem").Return(nil, expectedErr)

		result, err := t.Service.GetPublicKey(context.Background(), "pem")
		t.Error(err)
		t.Same(expectedErr, err)
		t.Empty(result)
	})
}

func TestLocalSignerService(t *testing.T) {
	suite.Run(t, new(LocalSignerServiceTestSuite))
}
