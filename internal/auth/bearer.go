package auth

import (
	"crypto/rand"
	"encoding/base64"
)

const (
	httpHeaderKey         = "Authorization"
	httpHeaderValuePrefix = "Bearer "
)

type BearerToken string

// NewBearerToken creates a simple unique base64 encoded bearer token.
func NewBearerToken() (BearerToken, error) {
	var token [16]byte

	_, err := rand.Read(token[:])
	if err != nil {
		return "", err
	}

	return BearerToken(base64.URLEncoding.EncodeToString(token[:])), nil
}
