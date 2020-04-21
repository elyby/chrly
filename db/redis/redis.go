package redis

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
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

var now = time.Now

func New(addr string, poolSize int) (*Redis, error) {
	conn, err := pool.New("tcp", addr, poolSize)
	if err != nil {
		return nil, err
	}

	return &Redis{
		pool: conn,
	}, nil
}

const accountIdToUsernameKey = "hash:username-to-account-id" // TODO: this should be actually "hash:user-id-to-username"
const mojangUsernameToUuidKey = "hash:mojang-username-to-uuid"

type Redis struct {
	pool *pool.Pool
}

func (db *Redis) FindSkinByUsername(username string) (*model.Skin, error) {
	conn, err := db.pool.Get()
	if err != nil {
		return nil, err
	}
	defer db.pool.Put(conn)

	return findByUsername(username, conn)
}

func findByUsername(username string, conn util.Cmder) (*model.Skin, error) {
	redisKey := buildUsernameKey(username)
	response := conn.Cmd("GET", redisKey)
	if response.IsType(redis.Nil) {
		return nil, nil
	}

	encodedResult, _ := response.Bytes()
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
	conn, err := db.pool.Get()
	if err != nil {
		return nil, err
	}
	defer db.pool.Put(conn)

	return findByUserId(id, conn)
}

func findByUserId(id int, conn util.Cmder) (*model.Skin, error) {
	response := conn.Cmd("HGET", accountIdToUsernameKey, id)
	if response.IsType(redis.Nil) {
		return nil, nil
	}

	username, err := response.Str()
	if err != nil {
		return nil, err
	}

	return findByUsername(username, conn)
}

func (db *Redis) SaveSkin(skin *model.Skin) error {
	conn, err := db.pool.Get()
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	return save(skin, conn)
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

func (db *Redis) RemoveSkinByUserId(id int) error {
	conn, err := db.pool.Get()
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	return removeByUserId(id, conn)
}

func removeByUserId(id int, conn util.Cmder) error {
	record, err := findByUserId(id, conn)
	if err != nil {
		return err
	}

	conn.Cmd("MULTI")

	conn.Cmd("HDEL", accountIdToUsernameKey, id)
	if record != nil {
		conn.Cmd("DEL", buildUsernameKey(record.Username))
	}

	conn.Cmd("EXEC")

	return nil
}

func (db *Redis) RemoveSkinByUsername(username string) error {
	conn, err := db.pool.Get()
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	return removeByUsername(username, conn)
}

func removeByUsername(username string, conn util.Cmder) error {
	record, err := findByUsername(username, conn)
	if err != nil {
		return err
	}

	if record == nil {
		return nil
	}

	conn.Cmd("MULTI")

	conn.Cmd("DEL", buildUsernameKey(record.Username))
	conn.Cmd("HDEL", accountIdToUsernameKey, record.UserId)

	conn.Cmd("EXEC")

	return nil
}

func (db *Redis) GetUuid(username string) (string, error) {
	conn, err := db.pool.Get()
	if err != nil {
		return "", err
	}
	defer db.pool.Put(conn)

	return findMojangUuidByUsername(username, conn)
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
	if storedAt.Add(time.Hour * 24 * 30).Before(now()) {
		return "", &mojangtextures.ValueNotFound{}
	}

	return parts[0], nil
}

func (db *Redis) StoreUuid(username string, uuid string) error {
	conn, err := db.pool.Get()
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	return storeMojangUuid(username, uuid, conn)
}

func storeMojangUuid(username string, uuid string, conn util.Cmder) error {
	value := uuid + ":" + strconv.FormatInt(now().Unix(), 10)
	res := conn.Cmd("HSET", mojangUsernameToUuidKey, strings.ToLower(username), value)
	if res.IsType(redis.Err) {
		return res.Err
	}

	return nil
}

func (db *Redis) Ping() error {
	r := db.pool.Cmd("PING")
	if r.Err != nil {
		return r.Err
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
	reader, err := zlib.NewReader(buff)
	if err != nil {
		return nil, err
	}

	resultBuffer := new(bytes.Buffer)
	_, _ = io.Copy(resultBuffer, reader)
	_ = reader.Close()

	return resultBuffer.Bytes(), nil
}
