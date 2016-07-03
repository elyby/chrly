package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mediocregopher/radix.v2/redis"
	"encoding/json"
	"strings"
	"time"
	"strconv"
	"crypto/md5"
	"encoding/hex"
)

var client, redisErr = redis.Dial("tcp", "redis:6379")

func main() {
	if redisErr != nil {
		log.Fatal("Redis unavailable")
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/skins/{username}", GetSkin)
	router.HandleFunc("/textures/{username}", GetTextures)

	log.Fatal(http.ListenAndServe(":80", router))
}

func GetSkin(w http.ResponseWriter, r *http.Request) {
	username := ParseUsername(mux.Vars(r)["username"])
	log.Println("request skin for username " + username);
	rec, err := FindRecord(username)
	if (err != nil) {
		http.Redirect(w, r, "http://skins.minecraft.net/MinecraftSkins/" + username + ".png", 301)
		log.Println("Cannot get skin for username " + username)
		return
	}

	http.Redirect(w, r, rec.Url, 301);
}

func GetTextures(w http.ResponseWriter, r *http.Request) {
	username := ParseUsername(mux.Vars(r)["username"])
	log.Println("request textures for username " + username)

	rec, err := FindRecord(username)
	if (err != nil || rec.SkinId == 0) {
		rec.Url = "http://skins.minecraft.net/MinecraftSkins/" + username + ".png"
		rec.Hash = string(BuildNonElyTexturesHash(username))
	}

	textures := TexturesResponse{
		Skin: &Skin{
			Url: rec.Url,
			Hash: rec.Hash,
		},
	}

	if (rec.IsSlim) {
		textures.Skin.Metadata = &SkinMetadata{
			Model: "slim",
		}
	}

	response,_ := json.Marshal(textures)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// STRUCTURES

type SkinItem struct {
	UserId   int    `json:"userId"`
	Nickname string `json:"nickname"`
	SkinId   int    `json:"skinId"`
	Url      string `json:"url"`
	Is1_8    bool   `json:"is1_8"`
	IsSlim   bool   `json:"isSlim"`
	Hash     string `json:"hash"`
}

type TexturesResponse struct {
	Skin *Skin `json:"SKIN"`
}

type Skin struct {
	Url      string `json:"url"`
	Hash     string `json:"hash"`
	Metadata *SkinMetadata `json:"metadata,omitempty"`
}

type SkinMetadata struct {
	Model string `json:"model"`
}

// TOOLS

func ParseUsername(username string) string {
	const suffix = ".png"
	if strings.HasSuffix(username, suffix) {
		username = strings.TrimSuffix(username, suffix)
	}

	return username
}

func BuildNonElyTexturesHash(username string) string {
	n := time.Now()
	hour := time.Date(n.Year(), n.Month(), n.Day(), n.Hour(), 0, 0, 0, time.UTC).Unix()
	hasher := md5.New()
	hasher.Write([]byte("non-ely-" + strconv.FormatInt(hour, 10) + "-" + username))

	return hex.EncodeToString(hasher.Sum(nil))
}

func FindRecord(username string) (SkinItem, error) {
	var record SkinItem;
	result, err := client.Cmd("GET", BuildKey(username)).Str();
	if (err == nil) {
		decodeErr := json.Unmarshal([]byte(result), &record)
		if (decodeErr != nil) {
			log.Println("Cannot decode record data")
		}
	}

	return record, err
}

func BuildKey(username string) string {
	return "username:" + strings.ToLower(username)
}
