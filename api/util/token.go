package util

import (
	"errors"
	"strings"
)

func ExtractBearerToken(authHeader string) (string, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("Authorization header must start with Bearer")
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}
