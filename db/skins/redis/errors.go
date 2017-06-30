package redis

import "fmt"

type SkinNotFound struct {
	Who string
}

func (e SkinNotFound) Error() string {
	return fmt.Sprintf("Skin data not found. Required username \"%v\"", e.Who)
}

