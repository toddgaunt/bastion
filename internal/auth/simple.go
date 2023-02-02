package auth

import (
	"time"

	"github.com/toddgaunt/bastion/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

// Simple is a basic single user in memory authenticator.
type Simple struct {
	username string
	hash     []byte
}

func NewSimple(username, password string) (*Simple, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Simple{
		username: username,
		hash:     hash,
	}, nil
}

func (sa *Simple) Authenticate(username, password string) (Claims, error) {
	// Don't allow either empty usernames or passwords
	if sa.username == "" || sa.hash == nil {
		return Claims{}, errors.New("invalid username and password")
	}

	err := bcrypt.CompareHashAndPassword(sa.hash, []byte(password))
	if err != nil {
		return Claims{}, err
	}

	if sa.username != username {
		return Claims{}, errors.New("invalid username and password")
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
