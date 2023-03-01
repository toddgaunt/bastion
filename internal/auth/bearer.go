package auth

import (
	"encoding/base64"
)

// BearerToken is a base64 encoded string.
type BearerToken string

// NewBearerToken creates a simple unique base64 encoded bearer token.
func NewBearerToken() (BearerToken, error) {
	token, err := ReadBytes(16)
	if err != nil {
		return "", err
	}

	return BearerToken(base64.URLEncoding.EncodeToString(token[:])), nil
}
