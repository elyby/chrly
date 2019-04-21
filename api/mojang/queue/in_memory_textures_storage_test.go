package queue

import (
	"time"

	"github.com/elyby/chrly/api/mojang"

	testify "github.com/stretchr/testify/assert"
	"testing"
)

var texturesWithSkin = &mojang.SignedTexturesResponse{
	Id:   "dead24f9a4fa4877b7b04c8c6c72bb46",
	Name: "mock",
	Props: []*mojang.Property{
		{
			Name: "textures",
			Value: mojang.EncodeTextures(&mojang.TexturesProp{
				Timestamp:   time.Now().UnixNano() / 10e5,
				ProfileID:   "dead24f9a4fa4877b7b04c8c6c72bb46",
				ProfileName: "mock",
				Textures: &mojang.TexturesResponse{
					Skin: &mojang.SkinTexturesResponse{
						Url: "http://textures.minecraft.net/texture/74d1e08b0bb7e9f590af27758125bbed1778ac6cef729aedfcb9613e9911ae75",
					},
				},
			}),
		},
	},
}
var texturesWithoutSkin = &mojang.SignedTexturesResponse{
	Id:   "dead24f9a4fa4877b7b04c8c6c72bb46",
	Name: "mock",
	Props: []*mojang.Property{
		{
			Name: "textures",
			Value: mojang.EncodeTextures(&mojang.TexturesProp{
				Timestamp:   time.Now().UnixNano() / 10e5,
				ProfileID:   "dead24f9a4fa4877b7b04c8c6c72bb46",
				ProfileName: "mock",
				Textures:    &mojang.TexturesResponse{},
			}),
		},
	},
}

func TestInMemoryTexturesStorage_GetTextures(t *testing.T) {
	t.Run("get error when uuid is not exists", func(t *testing.T) {
		assert := testify.New(t)

		storage := CreateInMemoryTexturesStorage()
		result, err := storage.GetTextures("b5d58475007d4f9e9ddd1403e2497579")

		assert.Nil(result)
		assert.Error(err, "value not found in the storage")
	})

	t.Run("get textures object, when uuid is stored in the storage", func(t *testing.T) {
		assert := testify.New(t)

		storage := CreateInMemoryTexturesStorage()
		storage.StoreTextures(texturesWithSkin)
		result, err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

		assert.Equal(texturesWithSkin, result)
		assert.Nil(err)
	})

	t.Run("get error when uuid is exists, but textures are expired", func(t *testing.T) {
		assert := testify.New(t)

		storage := CreateInMemoryTexturesStorage()
		storage.StoreTextures(texturesWithSkin)

		now = func() time.Time {
			return time.Now().Add(time.Minute * 2)
		}

		result, err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

		assert.Nil(result)
		assert.Error(err, "value not found in the storage")

		now = time.Now
	})
}

func TestInMemoryTexturesStorage_StoreTextures(t *testing.T) {
	t.Run("store textures for previously not existed uuid", func(t *testing.T) {
		assert := testify.New(t)

		storage := CreateInMemoryTexturesStorage()
		storage.StoreTextures(texturesWithSkin)
		result, err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

		assert.Equal(texturesWithSkin, result)
		assert.Nil(err)
	})

	t.Run("override already existed textures for uuid", func(t *testing.T) {
		assert := testify.New(t)

		storage := CreateInMemoryTexturesStorage()
		storage.StoreTextures(texturesWithoutSkin)
		storage.StoreTextures(texturesWithSkin)
		result, err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

		assert.NotEqual(texturesWithoutSkin, result)
		assert.Equal(texturesWithSkin, result)
		assert.Nil(err)
	})
}

func TestInMemoryTexturesStorage_GarbageCollection(t *testing.T) {
	assert := testify.New(t)

	inMemoryStorageGCPeriod = 10 * time.Millisecond
	inMemoryStoragePersistPeriod = 10 * time.Millisecond

	textures1 := &mojang.SignedTexturesResponse{
		Id:   "dead24f9a4fa4877b7b04c8c6c72bb46",
		Name: "mock1",
		Props: []*mojang.Property{
			{
				Name: "textures",
				Value: mojang.EncodeTextures(&mojang.TexturesProp{
					Timestamp:   time.Now().Add(inMemoryStorageGCPeriod-time.Millisecond*time.Duration(5)).UnixNano() / 10e5,
					ProfileID:   "dead24f9a4fa4877b7b04c8c6c72bb46",
					ProfileName: "mock1",
					Textures:    &mojang.TexturesResponse{},
				}),
			},
		},
	}
	textures2 := &mojang.SignedTexturesResponse{
		Id:   "b5d58475007d4f9e9ddd1403e2497579",
		Name: "mock2",
		Props: []*mojang.Property{
			{
				Name: "textures",
				Value: mojang.EncodeTextures(&mojang.TexturesProp{
					Timestamp:   time.Now().Add(inMemoryStorageGCPeriod-time.Millisecond*time.Duration(15)).UnixNano() / 10e5,
					ProfileID:   "b5d58475007d4f9e9ddd1403e2497579",
					ProfileName: "mock2",
					Textures:    &mojang.TexturesResponse{},
				}),
			},
		},
	}

	storage := CreateInMemoryTexturesStorage()
	storage.StoreTextures(textures1)
	storage.StoreTextures(textures2)

	storage.Start()

	time.Sleep(inMemoryStorageGCPeriod + time.Millisecond) // Let it start first iteration

	_, textures1Err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")
	_, textures2Err := storage.GetTextures("b5d58475007d4f9e9ddd1403e2497579")

	assert.Nil(textures1Err)
	assert.Error(textures2Err)

	time.Sleep(inMemoryStorageGCPeriod + time.Millisecond) // Let another iteration happen

	_, textures1Err = storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")
	_, textures2Err = storage.GetTextures("b5d58475007d4f9e9ddd1403e2497579")

	assert.Error(textures1Err)
	assert.Error(textures2Err)

	storage.Stop()
}
