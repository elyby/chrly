package db

type ParamRequired struct {
	Param string
}

func (e ParamRequired) Error() string {
	return "Required parameter not provided"
}
