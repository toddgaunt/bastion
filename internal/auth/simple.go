package auth

import (
	"time"

	"github.com/toddgaunt/bastion/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

// SimpleInitErr is returned when a Simple auth can't be initialized.
var SimpleInitErr = errors.New("username and password can't be empty string")

// simple is a basic single user in memory authenticator.
type simple struct {
	username string
	hash     []byte
}

func NewSimple(username, password string) (Authenticator, error) {
	if username == "" || password == "" {
		return nil, SimpleInitErr
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return simple{
		username: username,
		hash:     hash,
	}, nil
}

func (sa simple) Authenticate(username, password string) (Claims, error) {
	// Don't allow either empty usernames or passwords
	if sa.username == "" || sa.hash == nil {
		return Claims{}, errors.New("invalid username and password")
	}

	if sa.username != username {
		return Claims{}, errors.New("invalid username and password")
	}

	err := bcrypt.CompareHashAndPassword(sa.hash, []byte(password))
	if err != nil {
		return Claims{}, err
	}

	now := time.Now()

	return Claims{
		Username:  username,
		Admin:     true,
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Expiry:    now.Add(time.Hour * 24).Unix(),
	}, nil
}

type disabled struct{}

func NewDisabled() Authenticator {
	return disabled{}
}

func (da disabled) Authenticate(username, password string) (Claims, error) {
	return Claims{}, errors.New("authentication is disabled")
}
