//go:build redis

package redis

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/mediocregopher/radix/v4"
	"github.com/stretchr/testify/mock"
	assert "github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"ely.by/chrly/internal/db"
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

type MockProfileSerializer struct {
	mock.Mock
}

func (m *MockProfileSerializer) Serialize(profile *db.Profile) ([]byte, error) {
	args := m.Called(profile)

	return []byte(args.String(0)), args.Error(1)
}

func (m *MockProfileSerializer) Deserialize(value []byte) (*db.Profile, error) {
	args := m.Called(value)
	var result *db.Profile
	if casted, ok := args.Get(0).(*db.Profile); ok {
		result = casted
	}

	return result, args.Error(1)
}

func TestNew(t *testing.T) {
	t.Run("should connect", func(t *testing.T) {
		conn, err := New(context.Background(), &MockProfileSerializer{}, redisAddr, 12)
		assert.Nil(t, err)
		assert.NotNil(t, conn)
	})

	t.Run("should return error", func(t *testing.T) {
		conn, err := New(context.Background(), &MockProfileSerializer{}, "localhost:12345", 12) // Use localhost to avoid DNS resolution
		assert.Error(t, err)
		assert.Nil(t, conn)
	})
}

type redisTestSuite struct {
	suite.Suite

	Redis      *Redis
	Serializer *MockProfileSerializer

	cmd func(cmd string, args ...interface{}) string
}

func (s *redisTestSuite) SetupSuite() {
	s.Serializer = &MockProfileSerializer{}

	ctx := context.Background()
	conn, err := New(ctx, s.Serializer, redisAddr, 10)
	if err != nil {
		panic(fmt.Errorf("cannot establish connection to redis: %w", err))
	}

	s.Redis = conn
	s.cmd = func(cmd string, args ...interface{}) string {
		var result string
		err := s.Redis.client.Do(ctx, radix.FlatCmd(&result, cmd, args...))
		if err != nil {
			panic(err)
		}

		return result
	}
}

func (s *redisTestSuite) SetupSubTest() {
	// Cleanup database before each test
	s.cmd("FLUSHALL")
}

func (s *redisTestSuite) TearDownSubTest() {
	s.Serializer.AssertExpectations(s.T())
	for _, call := range s.Serializer.ExpectedCalls {
		call.Unset()
	}
}

func TestRedis(t *testing.T) {
	suite.Run(t, new(redisTestSuite))
}

func (s *redisTestSuite) TestFindProfileByUsername() {
	ctx := context.Background()
	s.Run("exists record", func() {
		serializedData := []byte("mock.exists.profile")
		expectedProfile := &db.Profile{}
		s.cmd("HSET", usernameToProfileKey, "mock", serializedData)
		s.Serializer.On("Deserialize", serializedData).Return(expectedProfile, nil)

		profile, err := s.Redis.FindProfileByUsername(ctx, "Mock")
		s.Require().NoError(err)
		s.Require().Same(expectedProfile, profile)
	})

	s.Run("not exists record", func() {
		profile, err := s.Redis.FindProfileByUsername(ctx, "Mock")
		s.Require().NoError(err)
		s.Require().Nil(profile)
	})

	s.Run("an error from serializer implementation", func() {
		expectedError := errors.New("mock error")
		s.cmd("HSET", usernameToProfileKey, "mock", "some-invalid-mock-data")
		s.Serializer.On("Deserialize", mock.Anything).Return(nil, expectedError)

		profile, err := s.Redis.FindProfileByUsername(ctx, "Mock")
		s.Require().Nil(profile)
		s.Require().ErrorIs(err, expectedError)
	})
}

func (s *redisTestSuite) TestSaveProfile() {
	ctx := context.Background()

	s.Run("save new entity", func() {
		profile := &db.Profile{
			Uuid:     "f57f36d5-4f50-4728-948a-42d5d80b18f3",
			Username: "Mock",
		}
		serializedProfile := "serialized-profile"
		s.Serializer.On("Serialize", profile).Return(serializedProfile, nil)

		s.cmd("HSET", usernameToProfileKey, "mock", serializedProfile)
		s.cmd("HSET", userUuidToUsernameKey, "f57f36d54f504728948a42d5d80b18f3", "mock")

		err := s.Redis.SaveProfile(ctx, profile)
		s.Require().NoError(err)

		uuidResp := s.cmd("HGET", userUuidToUsernameKey, "f57f36d54f504728948a42d5d80b18f3")
		s.Require().Equal("mock", uuidResp)

		profileResp := s.cmd("HGET", usernameToProfileKey, "mock")
		s.Require().Equal(serializedProfile, profileResp)
	})

	s.Run("update exists record with changed username", func() {
		newProfile := &db.Profile{
			Uuid:     "f57f36d5-4f50-4728-948a-42d5d80b18f3",
			Username: "NewMock",
		}
		serializedNewProfile := "serialized-new-profile"
		s.Serializer.On("Serialize", newProfile).Return(serializedNewProfile, nil)

		s.cmd("HSET", usernameToProfileKey, "mock", "serialized-old-profile")
		s.cmd("HSET", userUuidToUsernameKey, "f57f36d54f504728948a42d5d80b18f3", "mock")

		err := s.Redis.SaveProfile(ctx, newProfile)
		s.Require().NoError(err)

		uuidResp := s.cmd("HGET", userUuidToUsernameKey, "f57f36d54f504728948a42d5d80b18f3")
		s.Require().Equal("newmock", uuidResp)

		newProfileResp := s.cmd("HGET", usernameToProfileKey, "newmock")
		s.Require().Equal(serializedNewProfile, newProfileResp)

		oldProfileResp := s.cmd("HGET", usernameToProfileKey, "mock")
		s.Require().Empty(oldProfileResp)
	})
}

func (s *redisTestSuite) TestRemoveProfileByUuid() {
	ctx := context.Background()

	s.Run("exists record", func() {
		s.cmd("HSET", usernameToProfileKey, "mock", "serialized-profile")
		s.cmd("HSET", userUuidToUsernameKey, "f57f36d54f504728948a42d5d80b18f3", "mock")

		err := s.Redis.RemoveProfileByUuid(ctx, "f57f36d5-4f50-4728-948a-42d5d80b18f3")
		s.Require().NoError(err)

		uuidResp := s.cmd("HGET", userUuidToUsernameKey, "f57f36d54f504728948a42d5d80b18f3")
		s.Require().Empty(uuidResp)

		profileResp := s.cmd("HGET", usernameToProfileKey, "mock")
		s.Require().Empty(profileResp)
	})

	s.Run("uuid exists, username is missing", func() {
		s.cmd("HSET", userUuidToUsernameKey, "f57f36d54f504728948a42d5d80b18f3", "mock")

		err := s.Redis.RemoveProfileByUuid(ctx, "f57f36d5-4f50-4728-948a-42d5d80b18f3")
		s.Require().NoError(err)

		uuidResp := s.cmd("HGET", userUuidToUsernameKey, "f57f36d54f504728948a42d5d80b18f3")
		s.Require().Empty(uuidResp)
	})

	s.Run("uuid not exists", func() {
		err := s.Redis.RemoveProfileByUuid(ctx, "f57f36d5-4f50-4728-948a-42d5d80b18f3")
		s.Require().NoError(err)
	})
}

func (s *redisTestSuite) TestGetUuidForMojangUsername() {
	ctx := context.Background()

	s.Run("exists record", func() {
		s.cmd("SET", "mojang:uuid:mock", "MoCk:d3ca513eb3e14946b58047f2bd3530fd")

		uuid, username, err := s.Redis.GetUuidForMojangUsername(ctx, "Mock")
		s.Require().NoError(err)
		s.Require().Equal("MoCk", username)
		s.Require().Equal("d3ca513eb3e14946b58047f2bd3530fd", uuid)
	})

	s.Run("exists record with empty uuid value", func() {
		s.cmd("SET", "mojang:uuid:mock", "MoCk:")

		uuid, username, err := s.Redis.GetUuidForMojangUsername(ctx, "Mock")
		s.Require().NoError(err)
		s.Require().Equal("MoCk", username)
		s.Require().Empty(uuid)
	})

	s.Run("not exists record", func() {
		uuid, username, err := s.Redis.GetUuidForMojangUsername(ctx, "Mock")
		s.Require().NoError(err)
		s.Require().Empty(username)
		s.Require().Empty(uuid)
	})
}

func (s *redisTestSuite) TestStoreUuid() {
	ctx := context.Background()

	s.Run("store uuid", func() {
		err := s.Redis.StoreMojangUuid(ctx, "MoCk", "d3ca513eb3e14946b58047f2bd3530fd")
		s.Require().NoError(err)

		resp := s.cmd("GET", "mojang:uuid:mock")
		s.Require().Equal(resp, "MoCk:d3ca513eb3e14946b58047f2bd3530fd")
	})

	s.Run("store empty uuid", func() {
		err := s.Redis.StoreMojangUuid(ctx, "MoCk", "")
		s.Require().NoError(err)

		resp := s.cmd("GET", "mojang:uuid:mock")
		s.Require().Equal(resp, "MoCk:")
	})
}

func (s *redisTestSuite) TestPing() {
	err := s.Redis.Ping(context.Background())
	s.Require().Nil(err)
}
