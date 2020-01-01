package db

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"github.com/elyby/chrly/http"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/util"

	"github.com/elyby/chrly/model"
	"github.com/elyby/chrly/mojangtextures"
)

type RedisFactory struct {
	Host     string
	Port     int
	PoolSize int
	pool     *pool.Pool
}

func (f *RedisFactory) CreateSkinsRepository() (http.SkinsRepository, error) {
	return f.createInstance()
}

func (f *RedisFactory) CreateCapesRepository() (http.CapesRepository, error) {
	panic("capes repository not supported for this storage type")
}

func (f *RedisFactory) CreateMojangUuidsRepository() (mojangtextures.UuidsStorage, error) {
	return f.createInstance()
}

func (f *RedisFactory) createInstance() (*redisDb, error) {
	p, err := f.getPool()
	if err != nil {
		return nil, err
	}

	return &redisDb{p}, nil
}

func (f *RedisFactory) getPool() (*pool.Pool, error) {
	if f.pool == nil {
		if f.Host == "" {
			return nil, &ParamRequired{"host"}
		}

		if f.Port == 0 {
			return nil, &ParamRequired{"port"}
		}

		addr := fmt.Sprintf("%s:%d", f.Host, f.Port)
		conn, err := pool.New("tcp", addr, f.PoolSize)
		if err != nil {
			return nil, err
		}

		f.pool = conn
	}

	return f.pool, nil
}

type redisDb struct {
	pool *pool.Pool
}

const accountIdToUsernameKey = "hash:username-to-account-id"
const mojangUsernameToUuidKey = "hash:mojang-username-to-uuid"

func (db *redisDb) FindByUsername(username string) (*model.Skin, error) {
	conn, err := db.pool.Get()
	if err != nil {
		return nil, err
	}
	defer db.pool.Put(conn)

	return findByUsername(username, conn)
}

func (db *redisDb) FindByUserId(id int) (*model.Skin, error) {
	conn, err := db.pool.Get()
	if err != nil {
		return nil, err
	}
	defer db.pool.Put(conn)

	return findByUserId(id, conn)
}

func (db *redisDb) Save(skin *model.Skin) error {
	conn, err := db.pool.Get()
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	return save(skin, conn)
}

func (db *redisDb) RemoveByUserId(id int) error {
	conn, err := db.pool.Get()
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	return removeByUserId(id, conn)
}

func (db *redisDb) RemoveByUsername(username string) error {
	conn, err := db.pool.Get()
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	return removeByUsername(username, conn)
}

func (db *redisDb) GetUuid(username string) (string, error) {
	conn, err := db.pool.Get()
	if err != nil {
		return "", err
	}
	defer db.pool.Put(conn)

	return findMojangUuidByUsername(username, conn)
}

func (db *redisDb) StoreUuid(username string, uuid string) error {
	conn, err := db.pool.Get()
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	return storeMojangUuid(username, uuid, conn)
}

func findByUsername(username string, conn util.Cmder) (*model.Skin, error) {
	if username == "" {
		return nil, &http.SkinNotFoundError{username}
	}

	redisKey := buildUsernameKey(username)
	response := conn.Cmd("GET", redisKey)
	if !response.IsType(redis.Str) {
		return nil, &http.SkinNotFoundError{username}
	}

	encodedResult, err := response.Bytes()
	if err != nil {
		return nil, err
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

func findByUserId(id int, conn util.Cmder) (*model.Skin, error) {
	response := conn.Cmd("HGET", accountIdToUsernameKey, id)
	if !response.IsType(redis.Str) {
		return nil, &http.SkinNotFoundError{"unknown"}
	}

	username, _ := response.Str()

	return findByUsername(username, conn)
}

func removeByUserId(id int, conn util.Cmder) error {
	record, err := findByUserId(id, conn)
	if err != nil {
		if _, ok := err.(*http.SkinNotFoundError); !ok {
			return err
		}
	}

	conn.Cmd("MULTI")

	conn.Cmd("HDEL", accountIdToUsernameKey, id)
	if record != nil {
		conn.Cmd("DEL", buildUsernameKey(record.Username))
	}

	conn.Cmd("EXEC")

	return nil
}

func removeByUsername(username string, conn util.Cmder) error {
	record, err := findByUsername(username, conn)
	if err != nil {
		if _, ok := err.(*http.SkinNotFoundError); ok {
			return nil
		}

		return err
	}

	conn.Cmd("MULTI")

	conn.Cmd("DEL", buildUsernameKey(record.Username))
	conn.Cmd("HDEL", accountIdToUsernameKey, record.UserId)

	conn.Cmd("EXEC")

	return nil
}

func save(skin *model.Skin, conn util.Cmder) error {
	conn.Cmd("MULTI")

	// If user has changed username, then we must delete his old username record
	if skin.OldUsername != "" && skin.OldUsername != skin.Username {
		conn.Cmd("DEL", buildUsernameKey(skin.OldUsername))
	}

	// If this is a new record or if the user has changed username, we set the value in the hash table
	if skin.OldUsername != "" || skin.OldUsername != skin.Username {
		conn.Cmd("HSET", accountIdToUsernameKey, skin.UserId, skin.Username)
	}

	str, _ := json.Marshal(skin)
	conn.Cmd("SET", buildUsernameKey(skin.Username), zlibEncode(str))

	conn.Cmd("EXEC")

	skin.OldUsername = skin.Username

	return nil
}

func findMojangUuidByUsername(username string, conn util.Cmder) (string, error) {
	response := conn.Cmd("HGET", mojangUsernameToUuidKey, strings.ToLower(username))
	if response.IsType(redis.Nil) {
		return "", &mojangtextures.ValueNotFound{}
	}

	data, _ := response.Str()
	parts := strings.Split(data, ":")
	timestamp, _ := strconv.ParseInt(parts[1], 10, 64)
	storedAt := time.Unix(timestamp, 0)
	if storedAt.Add(time.Hour * 24 * 30).Before(time.Now()) {
		return "", &mojangtextures.ValueNotFound{}
	}

	return parts[0], nil
}

func storeMojangUuid(username string, uuid string, conn util.Cmder) error {
	value := uuid + ":" + strconv.FormatInt(time.Now().Unix(), 10)
	res := conn.Cmd("HSET", mojangUsernameToUuidKey, strings.ToLower(username), value)
	if res.IsType(redis.Err) {
		return res.Err
	}

	return nil
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
	reader, readError := zlib.NewReader(buff)
	if readError != nil {
		return nil, readError
	}

	resultBuffer := new(bytes.Buffer)
	_, _ = io.Copy(resultBuffer, reader)
	reader.Close()

	return resultBuffer.Bytes(), nil
}
