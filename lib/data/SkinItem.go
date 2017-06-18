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
	UserId          int    `json:"userId"`
	Uuid            string `json:"uuid"`
	Username        string `json:"username"`
	SkinId          int    `json:"skinId"`
	Url             string `json:"url"`
	Is1_8           bool   `json:"is1_8"`
	IsSlim          bool   `json:"isSlim"`
	Hash            string `json:"hash"`
	MojangTextures  string `json:"mojangTextures"`
	MojangSignature string `json:"mojangSignature"`
	oldUsername     string
}

const accountIdToUsernameKey string = "hash:username-to-account-id"

func (s *SkinItem) Save() {
	str, _ := json.Marshal(s)
	compressedStr := tools.ZlibEncode(str)
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

	pool.Cmd("SET", tools.BuildKey(s.Username), compressedStr)

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
	redisKey := tools.BuildKey(username)
	response := services.RedisPool.Cmd("GET", redisKey);
	if (response.IsType(redis.Nil)) {
		services.Logger.IncCounter("storage.not_found", 1)
		return record, SkinNotFound{username}
	}

	encodedResult, err := response.Bytes()
	if err == nil {
		services.Logger.IncCounter("storage.found", 1)
		result, err := tools.ZlibDecode(encodedResult)
		if err != nil {
			log.Println("Cannot uncompress zlib for key " + redisKey)
			goto finish
		}

		err = json.Unmarshal(result, &record)
		if err != nil {
			log.Println("Cannot decode record data for key" + redisKey)
			goto finish
		}

		record.oldUsername = record.Username
	}

	finish:

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
