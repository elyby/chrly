//go:build redis

package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/mediocregopher/radix/v4"
	assert "github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/model"
)

var redisAddr string

func init() {
	host := "localhost"
	port := 6379
	if os.Getenv("STORAGE_REDIS_HOST") != "" {
		host = os.Getenv("STORAGE_REDIS_HOST")
	}

	if os.Getenv("STORAGE_REDIS_PORT") != "" {
		port, _ = strconv.Atoi(os.Getenv("STORAGE_REDIS_PORT"))
	}

	redisAddr = fmt.Sprintf("%s:%d", host, port)
}

func TestNew(t *testing.T) {
	t.Run("should connect", func(t *testing.T) {
		conn, err := New(context.Background(), redisAddr, 12)
		assert.Nil(t, err)
		assert.NotNil(t, conn)
	})

	t.Run("should return error", func(t *testing.T) {
		conn, err := New(context.Background(), "localhost:12345", 12) // Use localhost to avoid DNS resolution
		assert.Error(t, err)
		assert.Nil(t, conn)
	})
}

type redisTestSuite struct {
	suite.Suite

	Redis *Redis

	cmd func(cmd string, args ...interface{}) string
}

func (suite *redisTestSuite) SetupSuite() {
	ctx := context.Background()
	conn, err := New(ctx, redisAddr, 10)
	if err != nil {
		panic(fmt.Errorf("cannot establish connection to redis: %w", err))
	}

	suite.Redis = conn
	suite.cmd = func(cmd string, args ...interface{}) string {
		var result string
		err := suite.Redis.client.Do(ctx, radix.FlatCmd(&result, cmd, args...))
		if err != nil {
			panic(err)
		}

		return result
	}
}

func (suite *redisTestSuite) SetupTest() {
	// Cleanup database before each test
	suite.cmd("FLUSHALL")
}

func (suite *redisTestSuite) TearDownTest() {
	// Restore time.Now func
	now = time.Now
}

func (suite *redisTestSuite) RunSubTest(name string, subTest func()) {
	suite.SetupTest()
	suite.Run(name, subTest)
}

func TestRedis(t *testing.T) {
	suite.Run(t, new(redisTestSuite))
}

/**
 * JSON with zlib encoding
 * {
 *     userId: 1,
 *     uuid: "fd5da1e4d66d4d17aadee2446093896d",
 *     username: "Mock",
 *     skinId: 1,
 *     url: "http://localhost/skin.png",
 *     is1_8: true,
 *     isSlim: false,
 *     mojangTextures: "mock-mojang-textures",
 *     mojangSignature: "mock-mojang-signature"
 * }
 */
var skinRecord = string([]byte{
	0x78, 0x9c, 0x5c, 0xce, 0x4b, 0x4a, 0x4, 0x41, 0xc, 0xc6, 0xf1, 0xbb, 0x7c, 0xeb, 0x1a, 0xdb, 0xd6, 0xb2,
	0x9c, 0xc9, 0xd, 0x5c, 0x88, 0x8b, 0xd1, 0xb5, 0x84, 0x4e, 0xa6, 0xa7, 0xec, 0x7a, 0xc, 0xf5, 0x0, 0x41,
	0xbc, 0xbb, 0xb4, 0xd2, 0xa, 0x2e, 0xf3, 0xe3, 0x9f, 0x90, 0xf, 0xf4, 0xaa, 0xe5, 0x41, 0x40, 0xa3, 0x41,
	0xef, 0x5e, 0x40, 0x38, 0xc9, 0x9d, 0xf0, 0xa8, 0x56, 0x9c, 0x13, 0x2b, 0xe3, 0x3d, 0xb3, 0xa8, 0xde, 0x58,
	0xeb, 0xae, 0xf, 0xb7, 0xfb, 0x83, 0x13, 0x98, 0xef, 0xa5, 0xc4, 0x51, 0x41, 0x78, 0xcc, 0xd3, 0x2, 0x83,
	0xba, 0xf8, 0xb4, 0x9d, 0x29, 0x1, 0x84, 0x73, 0x6b, 0x17, 0x1a, 0x86, 0x90, 0x27, 0xe, 0xe7, 0x5c, 0xdb,
	0xb0, 0x16, 0x57, 0x97, 0x34, 0xc3, 0xc0, 0xd7, 0xf1, 0x75, 0xf, 0x6a, 0xa5, 0xeb, 0x3a, 0x1c, 0x83, 0x8f,
	0xa0, 0x13, 0x87, 0xaa, 0x6, 0x31, 0xbf, 0x71, 0x9a, 0x9f, 0xf5, 0xbd, 0xf5, 0xa2, 0x15, 0x84, 0x98, 0xa7,
	0x65, 0xf7, 0xa3, 0xbb, 0xb6, 0xf1, 0xd6, 0x1d, 0xfd, 0x9c, 0x78, 0xa5, 0x7f, 0x61, 0xfd, 0x75, 0x83, 0xa7,
	0x20, 0x2f, 0x7f, 0xff, 0xe2, 0xf3, 0x2b, 0x0, 0x0, 0xff, 0xff, 0x6f, 0xdd, 0x51, 0x71,
})

func (suite *redisTestSuite) TestFindSkinByUsername() {
	suite.RunSubTest("exists record", func() {
		suite.cmd("SET", "username:mock", skinRecord)

		skin, err := suite.Redis.FindSkinByUsername("Mock")
		suite.Require().Nil(err)
		suite.Require().NotNil(skin)
		suite.Require().Equal(1, skin.UserId)
		suite.Require().Equal("fd5da1e4d66d4d17aadee2446093896d", skin.Uuid)
		suite.Require().Equal("Mock", skin.Username)
		suite.Require().Equal(1, skin.SkinId)
		suite.Require().Equal("http://localhost/skin.png", skin.Url)
		suite.Require().True(skin.Is1_8)
		suite.Require().False(skin.IsSlim)
		suite.Require().Equal("mock-mojang-textures", skin.MojangTextures)
		suite.Require().Equal("mock-mojang-signature", skin.MojangSignature)
		suite.Require().Equal(skin.Username, skin.OldUsername)
	})

	suite.RunSubTest("not exists record", func() {
		skin, err := suite.Redis.FindSkinByUsername("Mock")
		suite.Require().Nil(err)
		suite.Require().Nil(skin)
	})

	suite.RunSubTest("invalid zlib encoding", func() {
		suite.cmd("SET", "username:mock", "this is really not zlib")
		skin, err := suite.Redis.FindSkinByUsername("Mock")
		suite.Require().Nil(skin)
		suite.Require().EqualError(err, "zlib: invalid header")
	})

	suite.RunSubTest("invalid json encoding", func() {
		suite.cmd("SET", "username:mock", []byte{
			0x78, 0x9c, 0xca, 0x48, 0xcd, 0xc9, 0xc9, 0x57, 0x28, 0xcf, 0x2f, 0xca, 0x49, 0x1, 0x4, 0x0, 0x0, 0xff,
			0xff, 0x1a, 0xb, 0x4, 0x5d,
		})
		skin, err := suite.Redis.FindSkinByUsername("Mock")
		suite.Require().Nil(skin)
		suite.Require().EqualError(err, "invalid character 'h' looking for beginning of value")
	})
}

func (suite *redisTestSuite) TestFindSkinByUserId() {
	suite.RunSubTest("exists record", func() {
		suite.cmd("SET", "username:mock", skinRecord)
		suite.cmd("HSET", "hash:username-to-account-id", 1, "Mock")

		skin, err := suite.Redis.FindSkinByUserId(1)
		suite.Require().Nil(err)
		suite.Require().NotNil(skin)
		suite.Require().Equal(1, skin.UserId)
	})

	suite.RunSubTest("not exists record", func() {
		skin, err := suite.Redis.FindSkinByUserId(1)
		suite.Require().Nil(err)
		suite.Require().Nil(skin)
	})

	suite.RunSubTest("exists hash record, but no skin record", func() {
		suite.cmd("HSET", "hash:username-to-account-id", 1, "Mock")
		skin, err := suite.Redis.FindSkinByUserId(1)
		suite.Require().Nil(err)
		suite.Require().Nil(skin)
	})
}

func (suite *redisTestSuite) TestSaveSkin() {
	suite.RunSubTest("save new entity", func() {
		err := suite.Redis.SaveSkin(&model.Skin{
			UserId:          1,
			Uuid:            "fd5da1e4d66d4d17aadee2446093896d",
			Username:        "Mock",
			SkinId:          1,
			Url:             "http://localhost/skin.png",
			Is1_8:           true,
			IsSlim:          false,
			MojangTextures:  "mock-mojang-textures",
			MojangSignature: "mock-mojang-signature",
		})
		suite.Require().Nil(err)

		usernameResp := suite.cmd("GET", "username:mock")
		suite.Require().NotEmpty(usernameResp)
		suite.Require().Equal(skinRecord, usernameResp)

		idResp := suite.cmd("HGET", "hash:username-to-account-id", 1)
		suite.Require().Equal("Mock", idResp)
	})

	suite.RunSubTest("save exists record with changed username", func() {
		suite.cmd("SET", "username:mock", skinRecord)
		suite.cmd("HSET", "hash:username-to-account-id", 1, "Mock")

		err := suite.Redis.SaveSkin(&model.Skin{
			UserId:          1,
			Uuid:            "fd5da1e4d66d4d17aadee2446093896d",
			Username:        "NewMock",
			SkinId:          1,
			Url:             "http://localhost/skin.png",
			Is1_8:           true,
			IsSlim:          false,
			MojangTextures:  "mock-mojang-textures",
			MojangSignature: "mock-mojang-signature",
			OldUsername:     "Mock",
		})
		suite.Require().Nil(err)

		usernameResp := suite.cmd("GET", "username:newmock")
		suite.Require().NotEmpty(usernameResp)
		suite.Require().Equal(string([]byte{
			0x78, 0x9c, 0x5c, 0x8e, 0xcb, 0x4e, 0xc3, 0x40, 0xc, 0x45, 0xff, 0xe5, 0xae, 0xa7, 0x84, 0x40, 0x18, 0x5a,
			0xff, 0x1, 0xb, 0x60, 0x51, 0x58, 0x23, 0x2b, 0x76, 0xd3, 0x21, 0xf3, 0xa8, 0xe6, 0x21, 0x90, 0x10, 0xff,
			0x8e, 0x52, 0x14, 0x90, 0xba, 0xf4, 0xd1, 0xf1, 0xd5, 0xf9, 0x42, 0x2b, 0x9a, 0x1f, 0x4, 0xd4, 0x1b, 0xb4,
			0xe6, 0x4, 0x84, 0x83, 0xdc, 0x9, 0xf7, 0x3a, 0x88, 0xb5, 0x32, 0x48, 0x7f, 0xcf, 0x2c, 0xaa, 0x37, 0xc3,
			0x60, 0xaf, 0x77, 0xb7, 0xdb, 0x9d, 0x15, 0x98, 0xf3, 0x53, 0xe4, 0xa0, 0x20, 0x3c, 0xe9, 0xc7, 0x63, 0x1a,
			0x67, 0x18, 0x94, 0xd9, 0xc5, 0x75, 0x29, 0x7b, 0x10, 0x8e, 0xb5, 0x9e, 0xa8, 0xeb, 0x7c, 0x1a, 0xd9, 0x1f,
			0x53, 0xa9, 0xdd, 0x62, 0x5c, 0x9d, 0xe2, 0x4, 0x3, 0x57, 0xfa, 0xb7, 0x2d, 0xa8, 0xe6, 0xa6, 0xcb, 0xb1,
			0xf7, 0x2e, 0x80, 0xe, 0xec, 0x8b, 0x1a, 0x84, 0xf4, 0xce, 0x71, 0x7a, 0xd1, 0xcf, 0xda, 0xb2, 0x16, 0x10,
			0x42, 0x1a, 0xe7, 0xcd, 0x2f, 0xdd, 0xd4, 0x15, 0xaf, 0xde, 0xde, 0x4d, 0x91, 0x17, 0x74, 0x21, 0x96, 0x3f,
			0x6e, 0xf0, 0xec, 0xe5, 0xf5, 0x3f, 0xf9, 0xdc, 0xfb, 0xfd, 0x13, 0x0, 0x0, 0xff, 0xff, 0xca, 0xc3, 0x54,
			0x25,
		}), usernameResp)

		oldUsernameResp := suite.cmd("GET", "username:mock")
		suite.Require().Empty(oldUsernameResp)

		idResp := suite.cmd("HGET", "hash:username-to-account-id", 1)
		suite.Require().NotEmpty(usernameResp)
		suite.Require().Equal("NewMock", idResp)
	})
}

func (suite *redisTestSuite) TestRemoveSkinByUserId() {
	suite.RunSubTest("exists record", func() {
		suite.cmd("SET", "username:mock", skinRecord)
		suite.cmd("HSET", "hash:username-to-account-id", 1, "Mock")

		err := suite.Redis.RemoveSkinByUserId(1)
		suite.Require().Nil(err)

		usernameResp := suite.cmd("GET", "username:mock")
		suite.Require().Empty(usernameResp)

		idResp := suite.cmd("HGET", "hash:username-to-account-id", 1)
		suite.Require().Empty(idResp)
	})

	suite.RunSubTest("exists only id", func() {
		suite.cmd("HSET", "hash:username-to-account-id", 1, "Mock")

		err := suite.Redis.RemoveSkinByUserId(1)
		suite.Require().Nil(err)

		idResp := suite.cmd("HGET", "hash:username-to-account-id", 1)
		suite.Require().Empty(idResp)
	})

	suite.RunSubTest("error when querying skin record", func() {
		suite.cmd("SET", "username:mock", "invalid zlib")
		suite.cmd("HSET", "hash:username-to-account-id", 1, "Mock")

		err := suite.Redis.RemoveSkinByUserId(1)
		suite.Require().EqualError(err, "zlib: invalid header")
	})
}

func (suite *redisTestSuite) TestRemoveSkinByUsername() {
	suite.RunSubTest("exists record", func() {
		suite.cmd("SET", "username:mock", skinRecord)
		suite.cmd("HSET", "hash:username-to-account-id", 1, "Mock")

		err := suite.Redis.RemoveSkinByUsername("Mock")
		suite.Require().Nil(err)

		usernameResp := suite.cmd("GET", "username:mock")
		suite.Require().Empty(usernameResp)

		idResp := suite.cmd("HGET", "hash:username-to-account-id", 1)
		suite.Require().Empty(idResp)
	})

	suite.RunSubTest("exists only username", func() {
		suite.cmd("SET", "username:mock", skinRecord)

		err := suite.Redis.RemoveSkinByUsername("Mock")
		suite.Require().Nil(err)

		usernameResp := suite.cmd("GET", "username:mock")
		suite.Require().Empty(usernameResp)
	})

	suite.RunSubTest("no records", func() {
		err := suite.Redis.RemoveSkinByUsername("Mock")
		suite.Require().Nil(err)
	})

	suite.RunSubTest("error when querying skin record", func() {
		suite.cmd("SET", "username:mock", "invalid zlib")

		err := suite.Redis.RemoveSkinByUsername("Mock")
		suite.Require().EqualError(err, "zlib: invalid header")
	})
}

func (suite *redisTestSuite) TestGetUuid() {
	suite.RunSubTest("exists record", func() {
		suite.cmd("HSET",
			"hash:mojang-username-to-uuid",
			"mock",
			fmt.Sprintf("%s:%d", "d3ca513eb3e14946b58047f2bd3530fd", time.Now().Unix()),
		)

		uuid, found, err := suite.Redis.GetUuid("Mock")
		suite.Require().Nil(err)
		suite.Require().True(found)
		suite.Require().Equal("d3ca513eb3e14946b58047f2bd3530fd", uuid)
	})

	suite.RunSubTest("exists record with empty uuid value", func() {
		suite.cmd("HSET",
			"hash:mojang-username-to-uuid",
			"mock",
			fmt.Sprintf(":%d", time.Now().Unix()),
		)

		uuid, found, err := suite.Redis.GetUuid("Mock")
		suite.Require().Nil(err)
		suite.Require().True(found)
		suite.Require().Empty("", uuid)
	})

	suite.RunSubTest("not exists record", func() {
		uuid, found, err := suite.Redis.GetUuid("Mock")
		suite.Require().Nil(err)
		suite.Require().False(found)
		suite.Require().Empty(uuid)
	})

	suite.RunSubTest("exists, but expired record", func() {
		suite.cmd("HSET",
			"hash:mojang-username-to-uuid",
			"mock",
			fmt.Sprintf("%s:%d", "d3ca513eb3e14946b58047f2bd3530fd", time.Now().Add(-1*time.Hour*24*31).Unix()),
		)

		uuid, found, err := suite.Redis.GetUuid("Mock")
		suite.Require().Empty(uuid)
		suite.Require().False(found)
		suite.Require().Nil(err)

		resp := suite.cmd("HGET", "hash:mojang-username-to-uuid", "mock")
		suite.Require().Empty(resp, "should cleanup expired records")
	})

	suite.RunSubTest("exists, but corrupted record", func() {
		suite.cmd("HSET",
			"hash:mojang-username-to-uuid",
			"mock",
			"corrupted value",
		)

		uuid, found, err := suite.Redis.GetUuid("Mock")
		suite.Require().Empty(uuid)
		suite.Require().False(found)
		suite.Require().Error(err, "Got unexpected response from the mojangUsernameToUuid hash: \"corrupted value\"")

		resp := suite.cmd("HGET", "hash:mojang-username-to-uuid", "mock")
		suite.Require().Empty(resp, "should cleanup expired records")
	})
}

func (suite *redisTestSuite) TestStoreUuid() {
	suite.RunSubTest("store uuid", func() {
		now = func() time.Time {
			return time.Date(2020, 04, 21, 02, 10, 16, 0, time.UTC)
		}

		err := suite.Redis.StoreUuid("Mock", "d3ca513eb3e14946b58047f2bd3530fd")
		suite.Require().Nil(err)

		resp := suite.cmd("HGET", "hash:mojang-username-to-uuid", "mock")
		suite.Require().Equal(resp, "d3ca513eb3e14946b58047f2bd3530fd:1587435016")
	})

	suite.RunSubTest("store empty uuid", func() {
		now = func() time.Time {
			return time.Date(2020, 04, 21, 02, 10, 16, 0, time.UTC)
		}

		err := suite.Redis.StoreUuid("Mock", "")
		suite.Require().Nil(err)

		resp := suite.cmd("HGET", "hash:mojang-username-to-uuid", "mock")
		suite.Require().Equal(resp, ":1587435016")
	})
}

func (suite *redisTestSuite) TestPing() {
	err := suite.Redis.Ping()
	suite.Require().Nil(err)
}
