package auth

import (
	"crypto/rand"
	"encoding/json"
	"time"

	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/toddgaunt/bastion/internal/errors"
)

// JWT contains signed claim information.
type JWT string

// Claims contains information an authentication claims to verify.
type Claims struct {
	UserID string `json:"uid"`
	Expiry int64  `json:"exp"` // RFC 7519 4.1.4
}

const (
	keySize   = 16
	cryptAlgo = jose.A128GCM
	signAlgo  = jose.HS256
)

// Error keys exported by this package.
var (
	ErrJWTDecodeBytes = errors.Key("JWTDecode")
	ErrJWTBadAlgo     = errors.Key("JWTBadAlgo")
)

// NewClaims creates new claims with the information provided
func NewClaims(uid string, lifetime time.Duration) Claims {
	return Claims{
		UserID: uid,
		Expiry: time.Now().Add(lifetime).Unix(),
	}
}

// Encrypt signs then encrypts claim information into a JSON Web Token (JWT).
func Encrypt(claims Claims, secret []byte) (JWT, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", errors.Errorf("failed to encrypt claims: %w", err)
	}

	token, err := jose.EncryptBytes(payload, jose.DIR, cryptAlgo, secret, jose.Zip(jose.DEF))
	return JWT(token), err
}

// Decrypt decrypts claim information from an encrypted JWT.
func Decrypt(token JWT, secret []byte) (Claims, error) {
	authJSON, JWTHeader, err := jose.DecodeBytes(string(token), secret)
	if err != nil {
		return Claims{}, errors.E{
			Msg: "failed to decrypt token",
			Key: ErrJWTDecodeBytes,
			Err: err,
		}
	}

	// Verify that the token was encrypted using the right algorithm.
	algo, ok := JWTHeader["enc"].(string)
	if !ok || algo != cryptAlgo {
		return Claims{}, errors.E{
			Msg: "only the the A128GCM encryption algorithm is supported",
			Key: ErrJWTBadAlgo,
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
func Sign(claims Claims, secret []byte) (JWT, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", errors.Errorf("failed to encrypt claims: %w", err)
	}

	token, err := jose.SignBytes(payload, jose.HS256, secret)
	return JWT(token), err
}

// Verify verifies the signature of a signed token.
func Verify(token JWT, secret []byte) (Claims, error) {
	authJSON, JWTHeader, err := jose.DecodeBytes(string(token), secret)
	if err != nil {
		return Claims{}, errors.E{
			Msg: "failed to verify token",
			Key: ErrJWTDecodeBytes,
			Err: err,
		}
	}

	// Verify that the token was signed using the right algorithm.
	algo, ok := JWTHeader["alg"].(string)
	if !ok || algo != signAlgo {
		return Claims{}, errors.E{
			Msg: "only the the HS256 signature algorithm is supported",
			Key: ErrJWTBadAlgo,
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

// IsExpired returns true if the authentication's expiry time has been eclipsed
// by the current system time.
func (a Claims) IsExpired() bool {
	return a.Expiry <= time.Now().Unix()
}

// Equal checks two authentication data for equivalent fields, excluding expiry.
func (a Claims) Equal(b Claims) bool {
	return a.UserID == b.UserID
}

// GenerateKey creates a random key for signing bearer tokens. It will always
// always create a key for valid use by EncryptAuthenticationToken and
// DecryptAuthenticationToken.
func GenerateKey() ([]byte, error) {
	var key = make([]byte, keySize)
	var err error
	if _, err = rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}
