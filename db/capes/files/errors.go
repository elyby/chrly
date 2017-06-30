package files

import "fmt"

type CapeNotFound struct {
	Who string
}

func (e CapeNotFound) Error() string {
	return fmt.Sprintf("Cape file not found. Required username \"%v\"", e.Who)
}
