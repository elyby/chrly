package db

type ParamRequired struct {
	Param string
}

func (e ParamRequired) Error() string {
	return "Required parameter not provided"
}

type SkinNotFoundError struct {
	Who string
}

func (e SkinNotFoundError) Error() string {
	return "Skin data not found."
}

type CapeNotFoundError struct {
	Who string
}

func (e CapeNotFoundError) Error() string {
	return "Cape file not found."
}
