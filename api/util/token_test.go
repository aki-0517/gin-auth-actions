package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractBearerToken(t *testing.T) {
	authHeader := "Bearer 123"
	token, err := ExtractBearerToken(authHeader)
	assert.Nil(t, err)
	assert.Equal(t, token, "123")
	assert.Equal(t, authHeader, "Bearer "+token)
}
