package utils

import "time"

func UnixMillisecond(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
