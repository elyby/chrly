package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestJsonSerializer(t *testing.T) {
	var testCases = map[string]*struct {
		*Profile
		Serialized []byte
		Error      error
	}{
		"full structure": {
			Profile: &Profile{
				Uuid:            "f57f36d54f504728948a42d5d80b18f3",
				Username:        "mock-username",
				SkinUrl:         "https://example.com/skin.png",
				SkinModel:       "slim",
				CapeUrl:         "https://example.com/cape.png",
				MojangTextures:  "eyJ0aW1lc3RhbXAiOjE0ODYzMzcyNTQ4NzIsInByb2ZpbGVJZCI6ImM0ZjFlNTZmNjFkMTQwYTc4YzMyOGQ5MTY2ZWVmOWU3IiwicHJvZmlsZU5hbWUiOiJXaHlZb3VSZWFkVGhpcyIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS83Mzk1NmE4ZTY0ZWU2ZDhlYzY1NmFkYmI0NDA0ZjhlYmZmMzQxMWIwY2I5MGIzMWNiNDc2ZWNiOTk2ZDNiOCJ9fX0=",
				MojangSignature: "QH+1rlQJYk8tW+8WlSJnzxZZUL5RIkeOO33dq84cgNoxwCkzL95Zy5pbPMFhoiMXXablqXeqyNRZDQa+OewgDBSZxm0BmkNmwdTLzCPHgnlNYhwbO4sirg3hKjCZ82ORZ2q7VP2NQIwNvc3befiCakhDlMWUuhjxe7p/HKNtmKA7a/JjzmzwW7BWMv8b88ZaQaMaAc7puFQcu2E54G2Zk2kyv3T1Bm7bV4m7ymbL8McOmQc6Ph7C95/EyqIK1a5gRBUHPEFIEj0I06YKTHsCRFU1U/hJpk98xXHzHuULJobpajqYXuVJ8QEVgF8k8dn9VkS8BMbXcjzfbb6JJ36v7YIV6Rlt75wwTk2wr3C3P0ij55y0iXth1HjwcEKsg54n83d9w8yQbkUCiTpMbOqxTEOOS7G2O0ZDBJDXAKQ4n5qCiCXKZ4febv4+dWVQtgfZHnpGJUD3KdduDKslMePnECOXMjGSAOQou//yze2EkL2rBpJtAAiOtvBlm/aWnDZpij5cQk+pWmeHWZIf0LSSlsYRUWRDk/VKBvUTEAO9fqOxWqmSgQRUY2Ea56u0ZsBb4vEa1UY6mlJj3+PNZaWu5aP2E9Unh0DIawV96eW8eFQgenlNXHMmXd4aOra4sz2eeOnY53JnJP+eVE4cB1hlq8RA2mnwTtcy3lahzZonOWc=",
			},
			Serialized: []byte(`{"uuid":"f57f36d54f504728948a42d5d80b18f3","username":"mock-username","skinUrl":"https://example.com/skin.png","skinModel":"slim","capeUrl":"https://example.com/cape.png","mojangTextures":"eyJ0aW1lc3RhbXAiOjE0ODYzMzcyNTQ4NzIsInByb2ZpbGVJZCI6ImM0ZjFlNTZmNjFkMTQwYTc4YzMyOGQ5MTY2ZWVmOWU3IiwicHJvZmlsZU5hbWUiOiJXaHlZb3VSZWFkVGhpcyIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS83Mzk1NmE4ZTY0ZWU2ZDhlYzY1NmFkYmI0NDA0ZjhlYmZmMzQxMWIwY2I5MGIzMWNiNDc2ZWNiOTk2ZDNiOCJ9fX0=","mojangSignature":"QH+1rlQJYk8tW+8WlSJnzxZZUL5RIkeOO33dq84cgNoxwCkzL95Zy5pbPMFhoiMXXablqXeqyNRZDQa+OewgDBSZxm0BmkNmwdTLzCPHgnlNYhwbO4sirg3hKjCZ82ORZ2q7VP2NQIwNvc3befiCakhDlMWUuhjxe7p/HKNtmKA7a/JjzmzwW7BWMv8b88ZaQaMaAc7puFQcu2E54G2Zk2kyv3T1Bm7bV4m7ymbL8McOmQc6Ph7C95/EyqIK1a5gRBUHPEFIEj0I06YKTHsCRFU1U/hJpk98xXHzHuULJobpajqYXuVJ8QEVgF8k8dn9VkS8BMbXcjzfbb6JJ36v7YIV6Rlt75wwTk2wr3C3P0ij55y0iXth1HjwcEKsg54n83d9w8yQbkUCiTpMbOqxTEOOS7G2O0ZDBJDXAKQ4n5qCiCXKZ4febv4+dWVQtgfZHnpGJUD3KdduDKslMePnECOXMjGSAOQou//yze2EkL2rBpJtAAiOtvBlm/aWnDZpij5cQk+pWmeHWZIf0LSSlsYRUWRDk/VKBvUTEAO9fqOxWqmSgQRUY2Ea56u0ZsBb4vEa1UY6mlJj3+PNZaWu5aP2E9Unh0DIawV96eW8eFQgenlNXHMmXd4aOra4sz2eeOnY53JnJP+eVE4cB1hlq8RA2mnwTtcy3lahzZonOWc="}`),
		},
		"default skin model": {
			Profile: &Profile{
				Uuid:     "f57f36d54f504728948a42d5d80b18f3",
				Username: "mock-username",
				SkinUrl:  "https://example.com/skin.png",
			},
			Serialized: []byte(`{"uuid":"f57f36d54f504728948a42d5d80b18f3","username":"mock-username","skinUrl":"https://example.com/skin.png"}`),
		},
		"cape only": {
			Profile: &Profile{
				Uuid:     "f57f36d54f504728948a42d5d80b18f3",
				Username: "mock-username",
				CapeUrl:  "https://example.com/cape.png",
			},
			Serialized: []byte(`{"uuid":"f57f36d54f504728948a42d5d80b18f3","username":"mock-username","capeUrl":"https://example.com/cape.png"}`),
		},
		"minimal structure": {
			Profile: &Profile{
				Uuid:     "f57f36d54f504728948a42d5d80b18f3",
				Username: "mock-username",
			},
			Serialized: []byte(`{"uuid":"f57f36d54f504728948a42d5d80b18f3","username":"mock-username"}`),
		},
		"invalid json structure": {
			Serialized: []byte(`this is not json`),
			Error:      errors.New(`cannot parse JSON: unexpected value found: "this is not json"; unparsed tail: "this is not json"`),
		},
	}

	serializer := NewJsonSerializer()
	t.Run("Serialize", func(t *testing.T) {
		for n, c := range testCases {
			if c.Profile == nil {
				continue
			}

			t.Run(n, func(t *testing.T) {
				result, err := serializer.Serialize(c.Profile)
				require.NoError(t, err)
				require.Equal(t, c.Serialized, result)
			})
		}
	})

	t.Run("Deserialize", func(t *testing.T) {
		for n, c := range testCases {
			t.Run(n, func(t *testing.T) {
				result, err := serializer.Deserialize(c.Serialized)
				require.Equal(t, c.Error, err)
				require.Equal(t, c.Profile, result)
			})
		}
	})
}

type ProfileSerializerMock struct {
	mock.Mock
}

func (m *ProfileSerializerMock) Serialize(profile *Profile) ([]byte, error) {
	args := m.Called(profile)
	var result []byte
	if casted, ok := args.Get(0).([]byte); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *ProfileSerializerMock) Deserialize(value []byte) (*Profile, error) {
	args := m.Called(value)
	var result *Profile
	if casted, ok := args.Get(0).(*Profile); ok {
		result = casted
	}

	return result, args.Error(1)
}

func TestZlibEncoder(t *testing.T) {
	profile := &Profile{
		Uuid:     "f57f36d54f504728948a42d5d80b18f3",
		Username: "mock-username",
	}

	t.Run("Serialize", func(t *testing.T) {
		t.Run("successfully", func(t *testing.T) {
			serializer := &ProfileSerializerMock{}
			serializer.On("Serialize", profile).Return([]byte("serialized-string"), nil)
			encoder := NewZlibEncoder(serializer)

			result, err := encoder.Serialize(profile)
			require.NoError(t, err)
			require.Equal(t, []byte{0x78, 0x9c, 0x2a, 0x4e, 0x2d, 0xca, 0x4c, 0xcc, 0xc9, 0xac, 0x4a, 0x4d, 0xd1, 0x2d, 0x2e, 0x29, 0xca, 0xcc, 0x4b, 0x7, 0x4, 0x0, 0x0, 0xff, 0xff, 0x3e, 0xd8, 0x6, 0xf1}, result)
		})

		t.Run("handle error from serializer", func(t *testing.T) {
			expectedError := errors.New("mock error")
			serializer := &ProfileSerializerMock{}
			serializer.On("Serialize", profile).Return(nil, expectedError)
			encoder := NewZlibEncoder(serializer)

			result, err := encoder.Serialize(profile)
			require.Same(t, expectedError, err)
			require.Nil(t, result)
		})
	})

	t.Run("Deserialize", func(t *testing.T) {
		t.Run("successfully", func(t *testing.T) {
			serializer := &ProfileSerializerMock{}
			serializer.On("Deserialize", []byte("serialized-string")).Return(profile, nil)
			encoder := NewZlibEncoder(serializer)

			result, err := encoder.Deserialize([]byte{0x78, 0x9c, 0x2a, 0x4e, 0x2d, 0xca, 0x4c, 0xcc, 0xc9, 0xac, 0x4a, 0x4d, 0xd1, 0x2d, 0x2e, 0x29, 0xca, 0xcc, 0x4b, 0x7, 0x4, 0x0, 0x0, 0xff, 0xff, 0x3e, 0xd8, 0x6, 0xf1})
			require.NoError(t, err)
			require.Equal(t, profile, result)
		})

		t.Run("handle an error from deserializer", func(t *testing.T) {
			expectedError := errors.New("mock error")

			serializer := &ProfileSerializerMock{}
			serializer.On("Deserialize", []byte("serialized-string")).Return(nil, expectedError)
			encoder := NewZlibEncoder(serializer)

			result, err := encoder.Deserialize([]byte{0x78, 0x9c, 0x2a, 0x4e, 0x2d, 0xca, 0x4c, 0xcc, 0xc9, 0xac, 0x4a, 0x4d, 0xd1, 0x2d, 0x2e, 0x29, 0xca, 0xcc, 0x4b, 0x7, 0x4, 0x0, 0x0, 0xff, 0xff, 0x3e, 0xd8, 0x6, 0xf1})
			require.Same(t, expectedError, err)
			require.Nil(t, result)
		})

		t.Run("handle invalid zlib encoding", func(t *testing.T) {
			encoder := NewZlibEncoder(&ProfileSerializerMock{})

			result, err := encoder.Deserialize([]byte{0x6d, 0x6f, 0x63, 0x6b})
			require.ErrorContains(t, err, "invalid")
			require.Nil(t, result)
		})
	})
}

func BenchmarkFastJsonSerializer(b *testing.B) {
	profile := &Profile{
		Uuid:            "f57f36d54f504728948a42d5d80b18f3",
		Username:        "mock-username",
		SkinUrl:         "https://example.com/skin.png",
		SkinModel:       "slim",
		CapeUrl:         "https://example.com/cape.png",
		MojangTextures:  "eyJ0aW1lc3RhbXAiOjE0ODYzMzcyNTQ4NzIsInByb2ZpbGVJZCI6ImM0ZjFlNTZmNjFkMTQwYTc4YzMyOGQ5MTY2ZWVmOWU3IiwicHJvZmlsZU5hbWUiOiJXaHlZb3VSZWFkVGhpcyIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS83Mzk1NmE4ZTY0ZWU2ZDhlYzY1NmFkYmI0NDA0ZjhlYmZmMzQxMWIwY2I5MGIzMWNiNDc2ZWNiOTk2ZDNiOCJ9fX0=",
		MojangSignature: "QH+1rlQJYk8tW+8WlSJnzxZZUL5RIkeOO33dq84cgNoxwCkzL95Zy5pbPMFhoiMXXablqXeqyNRZDQa+OewgDBSZxm0BmkNmwdTLzCPHgnlNYhwbO4sirg3hKjCZ82ORZ2q7VP2NQIwNvc3befiCakhDlMWUuhjxe7p/HKNtmKA7a/JjzmzwW7BWMv8b88ZaQaMaAc7puFQcu2E54G2Zk2kyv3T1Bm7bV4m7ymbL8McOmQc6Ph7C95/EyqIK1a5gRBUHPEFIEj0I06YKTHsCRFU1U/hJpk98xXHzHuULJobpajqYXuVJ8QEVgF8k8dn9VkS8BMbXcjzfbb6JJ36v7YIV6Rlt75wwTk2wr3C3P0ij55y0iXth1HjwcEKsg54n83d9w8yQbkUCiTpMbOqxTEOOS7G2O0ZDBJDXAKQ4n5qCiCXKZ4febv4+dWVQtgfZHnpGJUD3KdduDKslMePnECOXMjGSAOQou//yze2EkL2rBpJtAAiOtvBlm/aWnDZpij5cQk+pWmeHWZIf0LSSlsYRUWRDk/VKBvUTEAO9fqOxWqmSgQRUY2Ea56u0ZsBb4vEa1UY6mlJj3+PNZaWu5aP2E9Unh0DIawV96eW8eFQgenlNXHMmXd4aOra4sz2eeOnY53JnJP+eVE4cB1hlq8RA2mnwTtcy3lahzZonOWc=",
	}
	serializedProfile := []byte(`{"uuid":"f57f36d54f504728948a42d5d80b18f3","username":"mock-username","skinUrl":"https://example.com/skin.png","skinModel":"slim","capeUrl":"https://example.com/cape.png","mojangTextures":"eyJ0aW1lc3RhbXAiOjE0ODYzMzcyNTQ4NzIsInByb2ZpbGVJZCI6ImM0ZjFlNTZmNjFkMTQwYTc4YzMyOGQ5MTY2ZWVmOWU3IiwicHJvZmlsZU5hbWUiOiJXaHlZb3VSZWFkVGhpcyIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS83Mzk1NmE4ZTY0ZWU2ZDhlYzY1NmFkYmI0NDA0ZjhlYmZmMzQxMWIwY2I5MGIzMWNiNDc2ZWNiOTk2ZDNiOCJ9fX0=","mojangSignature":"QH+1rlQJYk8tW+8WlSJnzxZZUL5RIkeOO33dq84cgNoxwCkzL95Zy5pbPMFhoiMXXablqXeqyNRZDQa+OewgDBSZxm0BmkNmwdTLzCPHgnlNYhwbO4sirg3hKjCZ82ORZ2q7VP2NQIwNvc3befiCakhDlMWUuhjxe7p/HKNtmKA7a/JjzmzwW7BWMv8b88ZaQaMaAc7puFQcu2E54G2Zk2kyv3T1Bm7bV4m7ymbL8McOmQc6Ph7C95/EyqIK1a5gRBUHPEFIEj0I06YKTHsCRFU1U/hJpk98xXHzHuULJobpajqYXuVJ8QEVgF8k8dn9VkS8BMbXcjzfbb6JJ36v7YIV6Rlt75wwTk2wr3C3P0ij55y0iXth1HjwcEKsg54n83d9w8yQbkUCiTpMbOqxTEOOS7G2O0ZDBJDXAKQ4n5qCiCXKZ4febv4+dWVQtgfZHnpGJUD3KdduDKslMePnECOXMjGSAOQou//yze2EkL2rBpJtAAiOtvBlm/aWnDZpij5cQk+pWmeHWZIf0LSSlsYRUWRDk/VKBvUTEAO9fqOxWqmSgQRUY2Ea56u0ZsBb4vEa1UY6mlJj3+PNZaWu5aP2E9Unh0DIawV96eW8eFQgenlNXHMmXd4aOra4sz2eeOnY53JnJP+eVE4cB1hlq8RA2mnwTtcy3lahzZonOWc="}`)

	serializer := NewJsonSerializer()
	b.Run("Serialize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = serializer.Serialize(profile)
		}
	})

	b.Run("Deserialize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = serializer.Deserialize(serializedProfile)
		}
	})
}
