package db

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/util"

	"elyby/minecraft-skinsystem/interfaces"
	"elyby/minecraft-skinsystem/model"
)

type RedisFactory struct {
	Host       string
	Port       int
	PoolSize   int
	connection *pool.Pool
}

// TODO: maybe we should manually return connection to the pool?

func (f RedisFactory) CreateSkinsRepository() (interfaces.SkinsRepository, error) {
	connection, err := f.getConnection()
	if err != nil {
		return nil, err
	}

	return &redisDb{connection}, nil
}

func (f RedisFactory) CreateCapesRepository() (interfaces.CapesRepository, error) {
	panic("capes repository not supported for this storage type")
}

func (f RedisFactory) getConnection() (*pool.Pool, error) {
	if f.connection == nil {
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

		f.connection = conn

		go func() {
			period := 5
			for {
				time.Sleep(time.Duration(period) * time.Second)
				resp := f.connection.Cmd("PING")
				if resp.Err == nil {
					continue
				}

				log.Println("Redis not pinged. Try to reconnect")
				conn, err := pool.New("tcp", addr, f.PoolSize)
				if err != nil {
					log.Printf("Cannot reconnect to redis: %v\n", err)
					log.Printf("Waiting %d seconds to retry\n", period)
					continue
				}

				f.connection = conn
				log.Println("Reconnected")
			}
		}()
	}

	return f.connection, nil
}

type redisDb struct {
	conn *pool.Pool
}

const accountIdToUsernameKey = "hash:username-to-account-id"

func (db *redisDb) FindByUsername(username string) (*model.Skin, error) {
	return findByUsername(username, db.getConn())
}

func (db *redisDb) FindByUserId(id int) (*model.Skin, error) {
	return findByUserId(id, db.getConn())
}

func (db *redisDb) Save(skin *model.Skin) error {
	return save(skin, db.getConn())
}

func (db *redisDb) RemoveByUserId(id int) error {
	return removeByUserId(id, db.getConn())
}

func (db *redisDb) RemoveByUsername(username string) error {
	return removeByUsername(username, db.getConn())
}

func (db *redisDb) getConn() util.Cmder {
	conn, _ := db.conn.Get()
	return conn
}

func findByUsername(username string, conn util.Cmder) (*model.Skin, error) {
	if username == "" {
		return nil, &SkinNotFoundError{username}
	}

	redisKey := buildUsernameKey(username)
	response := conn.Cmd("GET", redisKey)
	if response.IsType(redis.Nil) {
		return nil, &SkinNotFoundError{username}
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
	if response.IsType(redis.Nil) {
		return nil, &SkinNotFoundError{"unknown"}
	}

	username, _ := response.Str()

	return findByUsername(username, conn)
}

func removeByUserId(id int, conn util.Cmder) error {
	record, err := findByUserId(id, conn)
	if err != nil {
		if _, ok := err.(*SkinNotFoundError); !ok {
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
		if _, ok := err.(*SkinNotFoundError); !ok {
			return err
		}
	}

	conn.Cmd("MULTI")

	conn.Cmd("DEL", buildUsernameKey(record.Username))
	if record != nil {
		conn.Cmd("HDEL", accountIdToUsernameKey, record.UserId)
	}

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

func buildUsernameKey(username string) string {
	return "username:" + strings.ToLower(username)
}

func zlibEncode(str []byte) []byte {
	var buff bytes.Buffer
	writer := zlib.NewWriter(&buff)
	writer.Write(str)
	writer.Close()

	return buff.Bytes()
}

func zlibDecode(bts []byte) ([]byte, error) {
	buff := bytes.NewReader(bts)
	reader, readError := zlib.NewReader(buff)
	if readError != nil {
		return nil, readError
	}

	resultBuffer := new(bytes.Buffer)
	io.Copy(resultBuffer, reader)
	reader.Close()

	return resultBuffer.Bytes(), nil
}
