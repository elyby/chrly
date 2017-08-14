package db

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/util"

	"elyby/minecraft-skinsystem/model"
	"elyby/minecraft-skinsystem/repositories"
)

type RedisFactory struct {
	Host       string
	Port       int
	PoolSize   int
	connection util.Cmder
}

func (f RedisFactory) CreateSkinsRepository() (repositories.SkinsRepository, error) {
	connection, err := f.getConnection()
	if err != nil {
		return nil, err
	}

	return &redisDb{connection}, nil
}

func (f RedisFactory) CreateCapesRepository() (repositories.CapesRepository, error) {
	panic("capes repository not supported for this storage type")
}

func (f RedisFactory) getConnection() (util.Cmder, error) {
	if f.connection == nil {
		if f.Host == "" {
			return nil, &ParamRequired{"host"}
		}

		if f.Port == 0 {
			return nil, &ParamRequired{"port"}
		}

		var conn util.Cmder
		var err error
		addr := fmt.Sprintf("%s:%d", f.Host, f.Port)
		if f.PoolSize > 1 {
			conn, err = pool.New("tcp", addr, f.PoolSize)
		} else {
			conn, err = redis.Dial("tcp", addr)
		}

		if err != nil {
			return nil, err
		}

		f.connection = conn
	}

	return f.connection, nil
}

type redisDb struct {
	conn util.Cmder
}

const accountIdToUsernameKey string = "hash:username-to-account-id"

func (db *redisDb) FindByUsername(username string) (model.Skin, error) {
	var record model.Skin
	if username == "" {
		return record, &SkinNotFoundError{username}
	}

	redisKey := buildKey(username)
	response := db.conn.Cmd("GET", redisKey)
	if response.IsType(redis.Nil) {
		return record, &SkinNotFoundError{username}
	}

	encodedResult, err := response.Bytes()
	if err == nil {
		result, err := zlibDecode(encodedResult)
		if err != nil {
			log.Println("Cannot uncompress zlib for key " + redisKey) // TODO: replace with valid error
			return record, err
		}

		err = json.Unmarshal(result, &record)
		if err != nil {
			log.Println("Cannot decode record data for key" + redisKey) // TODO: replace with valid error
			return record, nil
		}

		record.OldUsername = record.Username
	}

	return record, nil
}

func (db *redisDb) FindByUserId(id int) (model.Skin, error) {
	response := db.conn.Cmd("HGET", accountIdToUsernameKey, id)
	if response.IsType(redis.Nil) {
		return model.Skin{}, SkinNotFoundError{"unknown"}
	}

	username, _ := response.Str()

	return db.FindByUsername(username)
}

func buildKey(username string) string {
	return "username:" + strings.ToLower(username)
}

//noinspection GoUnusedFunction
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
