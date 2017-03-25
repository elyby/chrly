package data

import (
	"log"
	"fmt"
	"encoding/json"

	"elyby/minecraft-skinsystem/lib/services"
	"elyby/minecraft-skinsystem/lib/tools"

	"github.com/mediocregopher/radix.v2/redis"
)

type SkinItem struct {
	UserId      int    `json:"userId"`
	Username    string `json:"username"`
	SkinId      int    `json:"skinId"`
	Url         string `json:"url"`
	Is1_8       bool   `json:"is1_8"`
	IsSlim      bool   `json:"isSlim"`
	Hash        string `json:"hash"`
	oldUsername string
}

const accountIdToUsernameKey string = "hash:username-to-account-id"

func (s *SkinItem) Save() {
	str, _ := json.Marshal(s)
	pool, _ := services.RedisPool.Get()
	pool.Cmd("MULTI")

	// Если пользователь сменил ник, то мы должны удать его ключ
	if (s.oldUsername != "" && s.oldUsername != s.Username) {
		pool.Cmd("DEL", tools.BuildKey(s.oldUsername))
	}

	// Если это новая запись или если пользователь сменил ник, то обновляем значение в хэш-таблице
	if (s.oldUsername != "" || s.oldUsername != s.Username) {
		pool.Cmd("HSET", accountIdToUsernameKey, s.UserId, s.Username)
	}

	pool.Cmd("SET", tools.BuildKey(s.Username), str)

	pool.Cmd("EXEC")

	s.oldUsername = s.Username
}

func (s *SkinItem) Delete() {
	if (s.oldUsername == "") {
		return;
	}

	pool, _ := services.RedisPool.Get()
	pool.Cmd("MULTI")

	pool.Cmd("DEL", tools.BuildKey(s.oldUsername))
	pool.Cmd("HDEL", accountIdToUsernameKey, s.UserId)

	pool.Cmd("EXEC")
}

func FindSkinByUsername(username string) (SkinItem, error) {
	var record SkinItem;
	services.Logger.IncCounter("storage.query", 1)
	response := services.RedisPool.Cmd("GET", tools.BuildKey(username));
	if (response.IsType(redis.Nil)) {
		services.Logger.IncCounter("storage.not_found", 1)
		return record, SkinNotFound{username}
	}

	result, err := response.Str()
	if (err == nil) {
		services.Logger.IncCounter("storage.found", 1)
		decodeErr := json.Unmarshal([]byte(result), &record)
		if (decodeErr != nil) {
			log.Println("Cannot decode record data")
		}

		record.oldUsername = record.Username
	}

	return record, err
}

func FindSkinById(id int) (SkinItem, error) {
	response := services.RedisPool.Cmd("HGET", accountIdToUsernameKey, id);
	if (response.IsType(redis.Nil)) {
		return SkinItem{}, SkinNotFound{"unknown"}
	}

	username, _ := response.Str()

	return FindSkinByUsername(username)
}

type SkinNotFound struct {
	Who string
}

func (e SkinNotFound) Error() string {
	return fmt.Sprintf("Skin data not found. Required username \"%v\"", e.Who)
}
