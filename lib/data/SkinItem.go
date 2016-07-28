package data

import (
	"log"
	"encoding/json"

	"elyby/minecraft-skinsystem/lib/services"
	"elyby/minecraft-skinsystem/lib/tools"
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
	services.Redis.Cmd("SET", tools.BuildKey(s.Username), str)
}

func FindRecord(username string) (SkinItem, error) {
	var record SkinItem;
	result, err := services.Redis.Cmd("GET", tools.BuildKey(username)).Str();
	if (err == nil) {
		decodeErr := json.Unmarshal([]byte(result), &record)
		if (decodeErr != nil) {
			log.Println("Cannot decode record data")
		}
	} else {
		log.Println("Error on request user data")
	}

	return record, err
}
