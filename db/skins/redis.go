package skins

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
	"log"
	"strings"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/util"

	"elyby/minecraft-skinsystem/model"
	"elyby/minecraft-skinsystem/repositories"
)

type RedisSkinsFactory struct {
	Addr string
	PollSize int
}

func (cfg *RedisSkinsFactory) Create() (repositories.SkinsRepository, error) {
	conn, err := pool.New("tcp", cfg.Addr, cfg.PollSize)
	if err != nil {
		return nil, err
	}

	// TODO: здесь можно запустить горутину по восстановлению соединения

	return &redisDb{conn: conn}, nil
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
			log.Println("Cannot uncompress zlib for key " + redisKey)
			return record, err
		}

		err = json.Unmarshal(result, &record)
		if err != nil {
			log.Println("Cannot decode record data for key" + redisKey)
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
