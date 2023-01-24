package auth

import (
	"crypto/rand"
	"encoding/json"
	"time"

	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/toddgaunt/bastion/internal/errors"
)

const (
	keySize   = 32
	cryptAlgo = jose.A256GCM
	signAlgo  = jose.HS256
)

// JWT contains signed claim information.
type JWT string

type SymmetricKey [keySize]byte

// Claims contains information an authentication claims to verify.
type Claims struct {
	UserID    string `json:"uid"`
	Expiry    int64  `json:"exp"` // RFC 7519 4.1.4
	NotBefore int64  `json:"nbf"` // RFC 7519 4.1.5
	IssuedAt  int64  `json:"iat"` // RFC 7519 4.1.6
}

// Error keys exported by this package.
var (
	ErrJWTDecodeBytes = errors.E{Key: "JWTDecode"}
	ErrJWTBadAlgo     = errors.E{Key: "JWTBadAlgo"}
)

// NewClaims creates new claims with the information provided
func NewClaims(uid string, now time.Time, lifetime time.Duration) Claims {
	return Claims{
		UserID:    uid,
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Expiry:    now.Add(lifetime).Unix(),
	}
}

// Encrypt signs then encrypts claim information into a JSON Web Token (JWT).
func Encrypt(claims Claims, key SymmetricKey) (JWT, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", errors.Errorf("failed to encrypt claims: %w", err)
	}

	token, err := jose.EncryptBytes(payload, jose.DIR, cryptAlgo, key[0:keySize], jose.Zip(jose.DEF))
	return JWT(token), err
}

// Decrypt decrypts claim information from an encrypted JWT.
func Decrypt(token JWT, key SymmetricKey) (Claims, error) {
	authJSON, JWTHeader, err := jose.DecodeBytes(string(token), key[0:keySize])
	if err != nil {
		return Claims{}, errors.E{
			Msg: "failed to decrypt token",
			Key: ErrJWTDecodeBytes.Key,
			Err: err,
		}
	}

	// Verify that the token was encrypted using the right algorithm.
	algo, ok := JWTHeader["enc"].(string)
	if !ok || algo != cryptAlgo {
		return Claims{}, errors.E{
			Msg: errors.Msgf("only the the %s encryption algorithm is supported", cryptAlgo),
			Key: ErrJWTBadAlgo.Key,
			Err: errors.Errorf("unsupported algorithm %s", algo),
		}
	}

	// Unmarshal the access token to get the permission data of the
	// connecting client.
	var claims = Claims{}
	if err = json.Unmarshal(authJSON, &claims); err != nil {
		return Claims{}, err
	}

	return claims, nil
}

// Sign signs claim information without encrypting it.
func Sign(claims Claims, key SymmetricKey) (JWT, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", errors.Errorf("failed to encrypt claims: %w", err)
	}

	token, err := jose.SignBytes(payload, jose.HS256, key[0:keySize])
	return JWT(token), err
}

// Verify verifies the signature of a signed token.
func Verify(token JWT, key SymmetricKey) (Claims, error) {
	authJSON, JWTHeader, err := jose.DecodeBytes(string(token), key[0:keySize])
	if err != nil {
		return Claims{}, errors.E{
			Msg: "failed to verify token",
			Key: ErrJWTDecodeBytes.Key,
			Err: err,
		}
	}

	// Verify that the token was signed using the right algorithm.
	algo, ok := JWTHeader["alg"].(string)
	if !ok || algo != signAlgo {
		return Claims{}, errors.E{
			Msg: errors.Msgf("only the the %s signature algorithm is supported", signAlgo),
			Key: ErrJWTBadAlgo.Key,
			Err: errors.Errorf("unsupported algorithm %s", algo),
		}
	}

	// Unmarshal the access token to get the permission data of the
	// connecting client.
	var claims = Claims{}
	if err = json.Unmarshal(authJSON, &claims); err != nil {
		return Claims{}, err
	}

	return claims, nil
}

// IsValid returns true if the authentication's expiry time has been eclipsed
// by the passed in time.
func (a Claims) IsValid(now time.Time) bool {
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