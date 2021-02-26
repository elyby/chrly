package utils

import (
	"time"

	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestUnixMillisecond(t *testing.T) {
	loc, _ := time.LoadLocation("CET")
	d := time.Date(2021, 02, 26, 00, 43, 57, 987654321, loc)

	assert.Equal(t, int64(1614296637987), UnixMillisecond(d))
}
