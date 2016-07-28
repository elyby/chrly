package tools

import (
	"strings"
	"time"
	"crypto/md5"
	"strconv"
	"encoding/hex"
)

func ParseUsername(username string) string {
	const suffix = ".png"
	if strings.HasSuffix(username, suffix) {
		username = strings.TrimSuffix(username, suffix)
	}

	return username
}

func BuildNonElyTexturesHash(username string) string {
	hour := getCurrentHour()
	hasher := md5.New()
	hasher.Write([]byte("non-ely-" + strconv.FormatInt(hour, 10) + "-" + username))

	return hex.EncodeToString(hasher.Sum(nil))
}

func BuildKey(username string) string {
	return "username:" + strings.ToLower(username)
}

func BuildElyUrl(route string) string {
	return "http://ely.by" + route
}

func getCurrentHour() int64 {
	n := time.Now()
	return time.Date(n.Year(), n.Month(), n.Day(), n.Hour(), 0, 0, 0, time.UTC).Unix()
}

