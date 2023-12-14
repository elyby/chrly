package redis

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/mediocregopher/radix/v4"

	"github.com/elyby/chrly/model"
)

var now = time.Now

func New(ctx context.Context, addr string, poolSize int) (*Redis, error) {
	client, err := (radix.PoolConfig{Size: poolSize}).New(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Redis{
		client:  client,
		context: ctx,
	}, nil
}

const accountIdToUsernameKey = "hash:username-to-account-id" // TODO: this should be actually "hash:user-id-to-username"
const mojangUsernameToUuidKey = "hash:mojang-username-to-uuid"

type Redis struct {
	client  radix.Client
	context context.Context
}

func (db *Redis) FindSkinByUsername(username string) (*model.Skin, error) {
	var skin *model.Skin
	err := db.client.Do(db.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		var err error
		skin, err = findByUsername(ctx, conn, username)

		return err
	}))

	return skin, err
}

func findByUsername(ctx context.Context, conn radix.Conn, username string) (*model.Skin, error) {
	redisKey := buildUsernameKey(username)
	var encodedResult []byte
	err := conn.Do(ctx, radix.Cmd(&encodedResult, "GET", redisKey))
	if err != nil {
		return nil, err
	}

	if len(encodedResult) == 0 {
		return nil, nil
	}

	result, err := zlibDecode(encodedResult)
	if err != nil {
		return nil, err
	}

	var skin *model.Skin
	err = json.Unmarshal(result, &skin)
	if err != nil {
		return nil, err
	}

	skin.OldUsername = skin.Username

	return skin, nil
}

func (db *Redis) FindSkinByUserId(id int) (*model.Skin, error) {
	var skin *model.Skin
	err := db.client.Do(db.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		var err error
		skin, err = findByUserId(ctx, conn, id)

		return err
	}))

	return skin, err
}

func findByUserId(ctx context.Context, conn radix.Conn, id int) (*model.Skin, error) {
	var username string
	err := conn.Do(ctx, radix.FlatCmd(&username, "HGET", accountIdToUsernameKey, id))
	if err != nil {
		return nil, err
	}

	if username == "" {
		return nil, nil
	}

	return findByUsername(ctx, conn, username)
}

func (db *Redis) SaveSkin(skin *model.Skin) error {
	return db.client.Do(db.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		return save(ctx, conn, skin)
	}))
}

func save(ctx context.Context, conn radix.Conn, skin *model.Skin) error {
	err := conn.Do(ctx, radix.Cmd(nil, "MULTI"))
	if err != nil {
		return err
	}

	// If user has changed username, then we must delete his old username record
	if skin.OldUsername != "" && skin.OldUsername != skin.Username {
		err = conn.Do(ctx, radix.Cmd(nil, "DEL", buildUsernameKey(skin.OldUsername)))
		if err != nil {
			return err
		}
	}

	// If this is a new record or if the user has changed username, we set the value in the hash table
	if skin.OldUsername != "" || skin.OldUsername != skin.Username {
		err = conn.Do(ctx, radix.FlatCmd(nil, "HSET", accountIdToUsernameKey, skin.UserId, skin.Username))
	}

	str, _ := json.Marshal(skin)
	err = conn.Do(ctx, radix.FlatCmd(nil, "SET", buildUsernameKey(skin.Username), zlibEncode(str)))
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.Cmd(nil, "EXEC"))
	if err != nil {
		return err
	}

	skin.OldUsername = skin.Username

	return nil
}

func (db *Redis) RemoveSkinByUserId(id int) error {
	return db.client.Do(db.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		return removeByUserId(ctx, conn, id)
	}))
}

func removeByUserId(ctx context.Context, conn radix.Conn, id int) error {
	record, err := findByUserId(ctx, conn, id)
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.Cmd(nil, "MULTI"))
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.FlatCmd(nil, "HDEL", accountIdToUsernameKey, id))
	if err != nil {
		return err
	}

	if record != nil {
		err = conn.Do(ctx, radix.Cmd(nil, "DEL", buildUsernameKey(record.Username)))
		if err != nil {
			return err
		}
	}

	return conn.Do(ctx, radix.Cmd(nil, "EXEC"))
}

func (db *Redis) RemoveSkinByUsername(username string) error {
	return db.client.Do(db.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		return removeByUsername(ctx, conn, username)
	}))
}

func removeByUsername(ctx context.Context, conn radix.Conn, username string) error {
	record, err := findByUsername(ctx, conn, username)
	if err != nil {
		return err
	}

	if record == nil {
		return nil
	}

	err = conn.Do(ctx, radix.Cmd(nil, "MULTI"))
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.Cmd(nil, "DEL", buildUsernameKey(record.Username)))
	if err != nil {
		return err
	}

	err = conn.Do(ctx, radix.FlatCmd(nil, "HDEL", accountIdToUsernameKey, record.UserId))
	if err != nil {
		return err
	}

	return conn.Do(ctx, radix.Cmd(nil, "EXEC"))
}

func (db *Redis) GetUuid(username string) (string, bool, error) {
	var uuid string
	var found bool
	err := db.client.Do(db.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		var err error
		uuid, found, err = findMojangUuidByUsername(ctx, conn, username)

		return err
	}))

	return uuid, found, err
}

func findMojangUuidByUsername(ctx context.Context, conn radix.Conn, username string) (string, bool, error) {
	key := strings.ToLower(username)
	var result string
	err := conn.Do(ctx, radix.Cmd(&result, "HGET", mojangUsernameToUuidKey, key))
	if err != nil {
		return "", false, err
	}

	if result == "" {
		return "", false, nil
	}

	parts := strings.Split(result, ":")
	// https://github.com/elyby/chrly/issues/28
	if len(parts) < 2 {
		err = conn.Do(ctx, radix.Cmd(nil, "HDEL", mojangUsernameToUuidKey, key))
		if err != nil {
			return "", false, err
		}

		return "", false, fmt.Errorf("got unexpected response from the mojangUsernameToUuid hash: \"%s\"", result)
	}

	timestamp, _ := strconv.ParseInt(parts[1], 10, 64)
	storedAt := time.Unix(timestamp, 0)
	if storedAt.Add(time.Hour * 24 * 30).Before(now()) {
		err = conn.Do(ctx, radix.Cmd(nil, "HDEL", mojangUsernameToUuidKey, key))
		if err != nil {
			return "", false, err
		}

		return "", false, nil
	}

	return parts[0], true, nil
}

func (db *Redis) StoreUuid(username string, uuid string) error {
	return db.client.Do(db.context, radix.WithConn("", func(ctx context.Context, conn radix.Conn) error {
		return storeMojangUuid(ctx, conn, username, uuid)
	}))
}

func storeMojangUuid(ctx context.Context, conn radix.Conn, username string, uuid string) error {
	value := uuid + ":" + strconv.FormatInt(now().Unix(), 10)
	err := conn.Do(ctx, radix.Cmd(nil, "HSET", mojangUsernameToUuidKey, strings.ToLower(username), value))
	if err != nil {
		return err
	}

	return nil
}

func (db *Redis) Ping() error {
	return db.client.Do(db.context, radix.Cmd(nil, "PING"))
}

func buildUsernameKey(username string) string {
	return "username:" + strings.ToLower(username)
}

func zlibEncode(str []byte) []byte {
	var buff bytes.Buffer
	writer := zlib.NewWriter(&buff)
	_, _ = writer.Write(str)
	_ = writer.Close()

	return buff.Bytes()
}

func zlibDecode(bts []byte) ([]byte, error) {
	buff := bytes.NewReader(bts)
	reader, err := zlib.NewReader(buff)
	if err != nil {
		return nil, err
	}

	resultBuffer := new(bytes.Buffer)
	_, _ = io.Copy(resultBuffer, reader)
	_ = reader.Close()

	return resultBuffer.Bytes(), nil
}
