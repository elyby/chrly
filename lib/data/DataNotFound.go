package data

import "fmt"

type DataNotFound struct {
	Who string
}

func (e DataNotFound) Error() string {
	return fmt.Sprintf("Skin data not found. Required username \"%v\"", e.Who)
}
