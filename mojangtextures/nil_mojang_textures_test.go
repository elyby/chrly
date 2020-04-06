package mojangtextures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilProvider_GetForUsername(t *testing.T) {
	provider := &NilProvider{}
	result, err := provider.GetForUsername("username")
	assert.Nil(t, result)
	assert.Nil(t, err)
}
