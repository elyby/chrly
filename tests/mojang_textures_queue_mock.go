package tests

import (
	"github.com/elyby/chrly/api/mojang"

	"github.com/stretchr/testify/mock"
)

type MojangTexturesQueueMock struct {
	mock.Mock
}

func (m *MojangTexturesQueueMock) GetTexturesForUsername(username string) chan *mojang.SignedTexturesResponse {
	args := m.Called(username)
	result := make(chan *mojang.SignedTexturesResponse)
	arg := args.Get(0)
	switch arg.(type) {
	case *mojang.SignedTexturesResponse:
		go func() {
			result <- arg.(*mojang.SignedTexturesResponse)
		}()
	case chan *mojang.SignedTexturesResponse:
		return arg.(chan *mojang.SignedTexturesResponse)
	case nil:
		go func() {
			result <- nil
		}()
	default:
		panic("unsupported return value")
	}

	return result
}
