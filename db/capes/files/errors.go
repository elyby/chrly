package files

type CapeNotFoundError struct {
	Who string
}

func (e CapeNotFoundError) Error() string {
	return "Cape file not found."
}
