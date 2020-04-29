package mojangtextures

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"

	"github.com/elyby/chrly/api/mojang"
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
	t.Run("should return nil, nil when textures are unavailable", func(t *testing.T) {
		storage := NewInMemoryTexturesStorage()
		result, err := storage.GetTextures("b5d58475007d4f9e9ddd1403e2497579")

		assert.Nil(t, result)
		assert.Nil(t, err)
	})

	t.Run("get textures object, when uuid is stored in the storage", func(t *testing.T) {
		storage := NewInMemoryTexturesStorage()
		storage.StoreTextures("dead24f9a4fa4877b7b04c8c6c72bb46", texturesWithSkin)
		result, err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

		assert.Equal(t, texturesWithSkin, result)
		assert.Nil(t, err)
	})

	t.Run("should return nil, nil when textures are exists, but cache duration is expired", func(t *testing.T) {
		storage := NewInMemoryTexturesStorage()
		storage.StoreTextures("dead24f9a4fa4877b7b04c8c6c72bb46", texturesWithSkin)

		now = func() time.Time {
			return time.Now().Add(time.Minute * 2)
		}

		result, err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

		assert.Nil(t, result)
		assert.Nil(t, err)

		now = time.Now
	})
}

func TestInMemoryTexturesStorage_StoreTextures(t *testing.T) {
	t.Run("store textures for previously not existed uuid", func(t *testing.T) {
		storage := NewInMemoryTexturesStorage()
		storage.StoreTextures("dead24f9a4fa4877b7b04c8c6c72bb46", texturesWithSkin)
		result, err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

		assert.Equal(t, texturesWithSkin, result)
		assert.Nil(t, err)
	})

	t.Run("override already existed textures for uuid", func(t *testing.T) {
		storage := NewInMemoryTexturesStorage()
		storage.StoreTextures("dead24f9a4fa4877b7b04c8c6c72bb46", texturesWithoutSkin)
		storage.StoreTextures("dead24f9a4fa4877b7b04c8c6c72bb46", texturesWithSkin)
		result, err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

		assert.NotEqual(t, texturesWithoutSkin, result)
		assert.Equal(t, texturesWithSkin, result)
		assert.Nil(t, err)
	})

	t.Run("store nil textures", func(t *testing.T) {
		storage := NewInMemoryTexturesStorage()
		storage.StoreTextures("dead24f9a4fa4877b7b04c8c6c72bb46", nil)
		result, err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")

		assert.Nil(t, result)
		assert.Nil(t, err)
	})

	t.Run("should panic if textures prop is not decoded", func(t *testing.T) {
		toStore := &mojang.SignedTexturesResponse{
			Id:   "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			Name: "mock",
			Props: []*mojang.Property{
				{
					Name:      "textures",
					Value:     "totally not base64 encoded json",
					Signature: "totally not base64 encoded signature",
				},
			},
		}

		assert.Panics(t, func() {
			storage := NewInMemoryTexturesStorage()
			storage.StoreTextures("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", toStore)
		})
	})
}

func TestInMemoryTexturesStorage_GarbageCollection(t *testing.T) {
	storage := NewInMemoryTexturesStorage()
	storage.GCPeriod = 10 * time.Millisecond
	storage.Duration = 10 * time.Millisecond

	textures1 := &mojang.SignedTexturesResponse{
		Id:   "dead24f9a4fa4877b7b04c8c6c72bb46",
		Name: "mock1",
		Props: []*mojang.Property{
			{
				Name: "textures",
				Value: mojang.EncodeTextures(&mojang.TexturesProp{
					Timestamp:   time.Now().Add(storage.GCPeriod-time.Millisecond*time.Duration(5)).UnixNano() / 10e5,
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
					Timestamp:   time.Now().Add(storage.GCPeriod-time.Millisecond*time.Duration(15)).UnixNano() / 10e5,
					ProfileID:   "b5d58475007d4f9e9ddd1403e2497579",
					ProfileName: "mock2",
					Textures:    &mojang.TexturesResponse{},
				}),
			},
		},
	}

	storage.StoreTextures("dead24f9a4fa4877b7b04c8c6c72bb46", textures1)
	storage.StoreTextures("b5d58475007d4f9e9ddd1403e2497579", textures2)

	storage.Start()
	defer storage.Stop()

	time.Sleep(storage.GCPeriod + time.Millisecond) // Let it start first iteration

	texturesFromStorage1, textures1Err := storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")
	texturesFromStorage2, textures2Err := storage.GetTextures("b5d58475007d4f9e9ddd1403e2497579")

	assert.NotNil(t, texturesFromStorage1)
	assert.Nil(t, textures1Err)
	assert.Nil(t, texturesFromStorage2)
	assert.Nil(t, textures2Err)

	time.Sleep(storage.GCPeriod + time.Millisecond) // Let another iteration happen

	texturesFromStorage1, textures1Err = storage.GetTextures("dead24f9a4fa4877b7b04c8c6c72bb46")
	texturesFromStorage2, textures2Err = storage.GetTextures("b5d58475007d4f9e9ddd1403e2497579")

	assert.Nil(t, texturesFromStorage1)
	assert.Nil(t, textures1Err)
	assert.Nil(t, texturesFromStorage2)
	assert.Nil(t, textures2Err)
}
