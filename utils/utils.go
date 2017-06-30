package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
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

func BuildElyUrl(route string) string {
	prefix := "http://ely.by"
	if !strings.HasPrefix(route, prefix) {
		route = prefix + route
	}

	return route
}

var timeNow = time.Now

func getCurrentHour() int64 {
	n := timeNow()
	return time.Date(n.Year(), n.Month(), n.Day(), n.Hour(), 0, 0, 0, time.UTC).Unix()
}

