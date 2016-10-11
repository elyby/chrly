package data

import (
	"io"
	"os"
	"fmt"
	"strings"
	"crypto/md5"
	"encoding/hex"

	"elyby/minecraft-skinsystem/lib/services"
)

type CapeItem struct {
	File *os.File
}

func FindCapeByUsername(username string) (CapeItem, error) {
	var record CapeItem
	file, err := os.Open(services.RootFolder + "/data/capes/" + strings.ToLower(username) + ".png")
	if (err != nil) {
		return record, CapeNotFound{username}
	}

	record.File = file

	return record, err
}

func (cape *CapeItem) CalculateHash() string {
	hasher := md5.New()
	io.Copy(hasher, cape.File)

	return hex.EncodeToString(hasher.Sum(nil))
}

type CapeNotFound struct {
	Who string
}

func (e CapeNotFound) Error() string {
	return fmt.Sprintf("Cape file not found. Required username \"%v\"", e.Who)
}
