package redis

import (
	"elyby/minecraft-skinsystem/model"

	"encoding/json"
	"log"

	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/util"
)

type redisDb struct {
	conn util.Cmder
}

const accountIdToUsernameKey string = "hash:username-to-account-id"

func (db *redisDb) FindByUsername(username string) (model.Skin, error) {
	var record model.Skin
	redisKey := buildKey(username)
	response := db.conn.Cmd("GET", redisKey)
	if response.IsType(redis.Nil) {
		return record, SkinNotFound{username}
	}

	encodedResult, err := response.Bytes()
	if err == nil {
		result, err := zlibDecode(encodedResult)
		if err != nil {
			log.Println("Cannot uncompress zlib for key " + redisKey)
			goto finish
		}

		err = json.Unmarshal(result, &record)
		if err != nil {
			log.Println("Cannot decode record data for key" + redisKey)
			goto finish
		}

		record.OldUsername = record.Username
	}

	finish:

	return record, err
}

func (db *redisDb) FindByUserId(id int) (model.Skin, error) {
	response := db.conn.Cmd("HGET", accountIdToUsernameKey, id)
	if response.IsType(redis.Nil) {
		return model.Skin{}, SkinNotFound{"unknown"}
	}

	username, _ := response.Str()

	return db.FindByUsername(username)
}
