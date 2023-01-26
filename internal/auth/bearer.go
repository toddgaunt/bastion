package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
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

// AddTokenAsBearer adds JWT to an http header as a bearer token.
func AddTokenAsBearer(header http.Header, token JWT) {
	header.Set(httpHeaderKey, httpHeaderValuePrefix+string(token))
}

// GetTokenFromBearer retrieves the bearer token from an HTTP header.
// If the header contains an invalid bearer token, an empty string is returned.
func GetTokenFromBearer(header http.Header) JWT {
	var bearer = header.Get(httpHeaderKey)
	if !strings.HasPrefix(strings.ToUpper(bearer), strings.ToUpper(httpHeaderValuePrefix)) {
		return ""
	}

	return JWT(bearer[len(httpHeaderValuePrefix):])
}
