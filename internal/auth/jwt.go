package auth

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/toddgaunt/bastion/internal/errors"
)

const (
	keySize  = 32
	signAlgo = jose.HS256
)

// JWT contains signed claim information.
type JWT string

// Symmetrickey contains bytes for encrption or signing operations
type SymmetricKey [keySize]byte

const (
	ErrJWTDecode  = errors.Type("jwt-decode")
	ErrJWTBadAlgo = errors.Type("jwt-bad-algo")
)

// Sign creates a new JWT from a set of claims with the provided time and duration.
func (claims Claims) Sign(now time.Time, lifetime time.Duration, key SymmetricKey) (JWT, error) {
	claims.IssuedAt = now.Unix()
	claims.NotBefore = now.Unix()
	claims.Expiry = now.Add(lifetime).Unix()

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", errors.Errorf("failed to sign claims: %w", err)
	}

	token, err := jose.SignBytes(payload, jose.HS256, key[0:keySize])
	return JWT(token), err
}

// Verify checks the signature of the token with the provided key.
func (token JWT) Verify(key SymmetricKey) (Claims, error) {
	authJSON, JWTHeader, err := jose.DecodeBytes(string(token), key[0:keySize])
	if err != nil {
		return Claims{}, errors.Annotation{
			WithType: ErrJWTDecode,
		}.Wrap(err)
	}

	// Verify that the token was signed using the right algorithm.
	algo, ok := JWTHeader["alg"].(string)
	if !ok || algo != signAlgo {
		return Claims{}, errors.Annotation{
			WithType:   ErrJWTBadAlgo,
			WithDetail: fmt.Sprintf("only the the %s signature algorithm is supported", signAlgo),
		}.Wrap(errors.Errorf("unsupported algorithm %s", algo))
	}

	// Unmarshal the access token to get the permission data of the
	// connecting client.
	var claims = Claims{}
	if err = json.Unmarshal(authJSON, &claims); err != nil {
		return Claims{}, err
	}

	return claims, nil
}

// IsExpired returns true if the authentication's expiry time has been eclipsed
// by the passed in time.
func (a Claims) IsExpired(now time.Time) bool {
	return a.NotBefore <= now.Unix() && a.Expiry <= now.Unix()
}

// GenerateSymmetricKey creates a random symmetric key for signing or encrypting claims.
func GenerateSymmetricKey() (SymmetricKey, error) {
	var key SymmetricKey
	if n, err := rand.Read(key[:]); n != keySize || err != nil {
		return key, err
	}

	return key, nil
}
