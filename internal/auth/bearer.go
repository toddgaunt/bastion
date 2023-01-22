package auth

import (
	"net/http"
	"strings"
)

const (
	httpHeaderKey         = "Authorization"
	httpHeaderValuePrefix = "Bearer "
)

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
