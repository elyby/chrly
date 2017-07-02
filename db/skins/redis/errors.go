package redis

type SkinNotFoundError struct {
	Who string
}

func (e SkinNotFoundError) Error() string {
	return "Skin data not found."
}
