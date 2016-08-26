package data

import (
	"log"
	"encoding/json"

	"elyby/minecraft-skinsystem/lib/services"
	"elyby/minecraft-skinsystem/lib/tools"

	"github.com/mediocregopher/radix.v2/redis"
)

type SkinItem struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	SkinId   int    `json:"skinId"`
	Url      string `json:"url"`
	Is1_8    bool   `json:"is1_8"`
	IsSlim   bool   `json:"isSlim"`
	Hash     string `json:"hash"`
}

func (s *SkinItem) Save() {
	str, _ := json.Marshal(s)
	services.RedisPool.Cmd("SET", tools.BuildKey(s.Username), str)
}

func FindRecord(username string) (SkinItem, error) {
	var record SkinItem;
	response := services.RedisPool.Cmd("GET", tools.BuildKey(username));
	if (response.IsType(redis.Nil)) {
		return record, DataNotFound{username}
	}

	result, err := response.Str()
	if (err == nil) {
		decodeErr := json.Unmarshal([]byte(result), &record)
		if (decodeErr != nil) {
			log.Println("Cannot decode record data")
		}
	}

	return record, err
}
