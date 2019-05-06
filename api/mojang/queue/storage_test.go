package queue

import (
	"github.com/elyby/chrly/api/mojang"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type uuidsStorageMock struct {
	mock.Mock
}

func (m *uuidsStorageMock) GetUuid(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *uuidsStorageMock) StoreUuid(username string, uuid string) error {
	m.Called(username, uuid)
	return nil
}

type texturesStorageMock struct {
	mock.Mock
}

func (m *texturesStorageMock) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	args := m.Called(uuid)
	var result *mojang.SignedTexturesResponse
	if casted, ok := args.Get(0).(*mojang.SignedTexturesResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *texturesStorageMock) StoreTextures(uuid string, textures *mojang.SignedTexturesResponse) {
	m.Called(uuid, textures)
}

func TestSplittedStorage(t *testing.T) {
	createMockedStorage := func() (*SplittedStorage, *uuidsStorageMock, *texturesStorageMock) {
		uuidsStorage := &uuidsStorageMock{}
		texturesStorage := &texturesStorageMock{}

		return &SplittedStorage{uuidsStorage, texturesStorage}, uuidsStorage, texturesStorage
	}

	t.Run("GetUuid", func(t *testing.T) {
		storage, uuidsMock, _ := createMockedStorage()
		uuidsMock.On("GetUuid", "username").Once().Return("find me", nil)
		result, err := storage.GetUuid("username")
		assert.Nil(t, err)
		assert.Equal(t, "find me", result)
		uuidsMock.AssertExpectations(t)
	})

	t.Run("StoreUuid", func(t *testing.T) {
		storage, uuidsMock, _ := createMockedStorage()
		uuidsMock.On("StoreUuid", "username", "result").Once()
		_ = storage.StoreUuid("username", "result")
		uuidsMock.AssertExpectations(t)
	})

	t.Run("GetTextures", func(t *testing.T) {
		result := &mojang.SignedTexturesResponse{Id: "mock id"}
		storage, _, texturesMock := createMockedStorage()
		texturesMock.On("GetTextures", "uuid").Once().Return(result, nil)
		returned, err := storage.GetTextures("uuid")
		assert.Nil(t, err)
		assert.Equal(t, result, returned)
		texturesMock.AssertExpectations(t)
	})

	t.Run("StoreTextures", func(t *testing.T) {
		toStore := &mojang.SignedTexturesResponse{}
		storage, _, texturesMock := createMockedStorage()
		texturesMock.On("StoreTextures", "mock id", toStore).Once()
		storage.StoreTextures("mock id", toStore)
		texturesMock.AssertExpectations(t)
	})
}

func TestValueNotFound_Error(t *testing.T) {
	err := &ValueNotFound{}
	assert.Equal(t, "value not found in the storage", err.Error())
}
