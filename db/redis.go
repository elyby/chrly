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

	"elyby/minecraft-skinsystem/model"
	"elyby/minecraft-skinsystem/interfaces"
)

type RedisFactory struct {
	Host       string
	Port       int
	PoolSize   int
	connection util.Cmder
}

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

func (f RedisFactory) getConnection() (util.Cmder, error) {
	if f.connection == nil {
		if f.Host == "" {
			return nil, &ParamRequired{"host"}
		}

		if f.Port == 0 {
			return nil, &ParamRequired{"port"}
		}

		addr := fmt.Sprintf("%s:%d", f.Host, f.Port)
		conn, err := createConnection(addr, f.PoolSize)
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
				conn, err := createConnection(addr, f.PoolSize)
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

func createConnection(addr string, poolSize int) (util.Cmder, error) {
	if poolSize > 1 {
		return pool.New("tcp", addr, poolSize)
	} else {
		return redis.Dial("tcp", addr)
	}
}

type redisDb struct {
	conn util.Cmder
}

const accountIdToUsernameKey string = "hash:username-to-account-id"

func (db *redisDb) FindByUsername(username string) (*model.Skin, error) {
	if username == "" {
		return nil, &SkinNotFoundError{username}
	}

	redisKey := buildKey(username)
	response := db.conn.Cmd("GET", redisKey)
	if response.IsType(redis.Nil) {
		return nil, &SkinNotFoundError{username}
	}

	encodedResult, err := response.Bytes()
	if err != nil {
		return nil, err
	}

	result, err := zlibDecode(encodedResult)
	if err != nil {
		log.Println("Cannot uncompress zlib for key " + redisKey) // TODO: replace with valid error
		return nil, err
	}

	var skin *model.Skin
	err = json.Unmarshal(result, &skin)
	if err != nil {
		log.Println("Cannot decode record data for key" + redisKey) // TODO: replace with valid error
		return nil, nil
	}

	skin.OldUsername = skin.Username

	return skin, nil
}

func (db *redisDb) FindByUserId(id int) (*model.Skin, error) {
	response := db.conn.Cmd("HGET", accountIdToUsernameKey, id)
	if response.IsType(redis.Nil) {
		return nil, SkinNotFoundError{"unknown"}
	}

	username, _ := response.Str()

	return db.FindByUsername(username)
}

func (db *redisDb) Save(skin *model.Skin) error {
	conn := db.conn
	if poolConn, isPool := conn.(*pool.Pool); isPool {
		conn, _ = poolConn.Get()
	}

	conn.Cmd("MULTI")

	// Если пользователь сменил ник, то мы должны удать его ключ
	if skin.OldUsername != "" && skin.OldUsername != skin.Username {
		conn.Cmd("DEL", buildKey(skin.OldUsername))
	}

	// Если это новая запись или если пользователь сменил ник, то обновляем значение в хэш-таблице
	if skin.OldUsername != "" || skin.OldUsername != skin.Username {
		conn.Cmd("HSET", accountIdToUsernameKey, skin.UserId, skin.Username)
	}

	str, _ := json.Marshal(skin)
	conn.Cmd("SET", buildKey(skin.Username), zlibEncode(str))

	conn.Cmd("EXEC")

	skin.OldUsername = skin.Username

	return nil
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
