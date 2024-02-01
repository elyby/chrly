package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/mediocregopher/radix/v4"

	"github.com/elyby/chrly/internal/db"
)

const usernameToProfileKey = "hash:username-to-profile"
const userUuidToUsernameKey = "hash:uuid-to-username"

type Redis struct {
	client     radix.Client
	context    context.Context
	serializer db.ProfileSerializer
}

func New(ctx context.Context, profileSerializer db.ProfileSerializer, addr string, poolSize int) (*Redis, error) {
	client, err := (radix.PoolConfig{Size: poolSize}).New(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Redis{
		client:     client,
		context:    ctx,
		serializer: profileSerializer,
	}, nil
}

func (r *Redis) FindProfileByUsername(username string) (*db.Profile, error) {
	var profile *db.Profile
	err := r.client.Do(r.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		var err error
		profile, err = r.findProfileByUsername(ctx, conn, username)

		return err
	}))

	return profile, err
}

func (r *Redis) findProfileByUsername(ctx context.Context, conn radix.Conn, username string) (*db.Profile, error) {
	var encodedResult []byte
	err := conn.Do(ctx, radix.Cmd(&encodedResult, "HGET", usernameToProfileKey, usernameHashKey(username)))
	if err != nil {
		return nil, err
	}

	if len(encodedResult) == 0 {
		return nil, nil
	}

	return r.serializer.Deserialize(encodedResult)
}

func (r *Redis) FindProfileByUuid(uuid string) (*db.Profile, error) {
	var skin *db.Profile
	err := r.client.Do(r.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		var err error
		skin, err = r.findProfileByUuid(ctx, conn, uuid)

		return err
	}))

	return skin, err
}

func (r *Redis) findProfileByUuid(ctx context.Context, conn radix.Conn, uuid string) (*db.Profile, error) {
	username, err := r.findUsernameHashKeyByUuid(ctx, conn, uuid)
	if err != nil {
		return nil, err
	}

	if username == "" {
		return nil, nil
	}

	return r.findProfileByUsername(ctx, conn, username)
}

func (r *Redis) findUsernameHashKeyByUuid(ctx context.Context, conn radix.Conn, uuid string) (string, error) {
	var username string
	return username, conn.Do(ctx, radix.FlatCmd(&username, "HGET", userUuidToUsernameKey, normalizeUuid(uuid)))
}

func (r *Redis) SaveProfile(profile *db.Profile) error {
	return r.client.Do(r.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		return r.saveProfile(ctx, conn, profile)
	}))
}

func (r *Redis) saveProfile(ctx context.Context, conn radix.Conn, profile *db.Profile) error {
	newUsernameHashKey := usernameHashKey(profile.Username)
	existsUsernameHashKey, err := r.findUsernameHashKeyByUuid(ctx, conn, profile.Uuid)
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.Cmd(nil, "MULTI"))
	if err != nil {
		return err
	}

	// If user has changed username, then we must delete his old username record
	if existsUsernameHashKey != "" && existsUsernameHashKey != newUsernameHashKey {
		err = conn.Do(ctx, radix.Cmd(nil, "HDEL", usernameToProfileKey, existsUsernameHashKey))
		if err != nil {
			return err
		}
	}

	err = conn.Do(ctx, radix.FlatCmd(nil, "HSET", userUuidToUsernameKey, normalizeUuid(profile.Uuid), newUsernameHashKey))
	if err != nil {
		return err
	}

	serializedProfile, err := r.serializer.Serialize(profile)
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.FlatCmd(nil, "HSET", usernameToProfileKey, newUsernameHashKey, serializedProfile))
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.Cmd(nil, "EXEC"))
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) RemoveProfileByUuid(uuid string) error {
	return r.client.Do(r.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		return r.removeProfileByUuid(ctx, conn, uuid)
	}))
}

func (r *Redis) removeProfileByUuid(ctx context.Context, conn radix.Conn, uuid string) error {
	username, err := r.findUsernameHashKeyByUuid(ctx, conn, uuid)
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.Cmd(nil, "MULTI"))
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.FlatCmd(nil, "HDEL", userUuidToUsernameKey, normalizeUuid(uuid)))
	if err != nil {
		return err
	}

	if username != "" {
		err = conn.Do(ctx, radix.Cmd(nil, "HDEL", usernameToProfileKey, usernameHashKey(username)))
		if err != nil {
			return err
		}
	}

	return conn.Do(ctx, radix.Cmd(nil, "EXEC"))
}

func (r *Redis) GetUuidForMojangUsername(username string) (string, string, error) {
	var uuid string
	foundUsername := username
	err := r.client.Do(r.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		var err error
		uuid, foundUsername, err = findMojangUuidByUsername(ctx, conn, username)

		return err
	}))

	return uuid, foundUsername, err
}

func findMojangUuidByUsername(ctx context.Context, conn radix.Conn, username string) (string, string, error) {
	key := buildMojangUsernameKey(username)
	var result string
	err := conn.Do(ctx, radix.Cmd(&result, "GET", key))
	if err != nil {
		return "", "", err
	}

	if result == "" {
		return "", "", nil
	}

	parts := strings.Split(result, ":")

	return parts[1], parts[0], nil
}

func (r *Redis) StoreMojangUuid(username string, uuid string) error {
	return r.client.Do(r.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		return storeMojangUuid(ctx, conn, username, uuid)
	}))
}

func storeMojangUuid(ctx context.Context, conn radix.Conn, username string, uuid string) error {
	value := fmt.Sprintf("%s:%s", username, uuid)
	err := conn.Do(ctx, radix.FlatCmd(nil, "SET", buildMojangUsernameKey(username), value, "EX", 60*60*24*30))
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) Ping() error {
	return r.client.Do(r.context, radix.Cmd(nil, "PING"))
}

func normalizeUuid(uuid string) string {
	return strings.ToLower(strings.ReplaceAll(uuid, "-", ""))
}

func usernameHashKey(username string) string {
	return strings.ToLower(username)
}

func buildMojangUsernameKey(username string) string {
	return fmt.Sprintf("mojang:uuid:%s", usernameHashKey(username))
}
