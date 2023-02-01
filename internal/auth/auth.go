package auth

import (
	"crypto/rand"
)

type Authenticator interface {
	Authenticate(username, password string) (Claims, error)
}

// Claims contains information an authentication claims to verify.
type Claims struct {
	Username string `json:"uid,omitempty"`
	Admin    bool   `json:"adm,omitempty"`

	// These fields are filled in automatically.
	Expiry    int64 `json:"exp"` // RFC 7519 4.1.4
	NotBefore int64 `json:"nbf"` // RFC 7519 4.1.5
	IssuedAt  int64 `json:"iat"` // RFC 7519 4.1.6
}

// ReadBytes reads size bytes from /dev/urandom.
func ReadBytes(size int) ([]byte, error) {
	key := make([]byte, size)
	if n, err := rand.Read(key[:]); n != size || err != nil {
		return key, err
	}

	return key, nil
}
