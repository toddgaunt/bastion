package auth

import (
	"time"

	"github.com/toddgaunt/bastion/internal/errors"
)

type Simple struct {
	Username string
	Password string
}

func (sa Simple) Authenticate(username, password string) (Claims, error) {
	// Don't allow either empty usernames or passwords
	if sa.Username == "" || sa.Password == "" {
		return Claims{}, errors.New("invalid username and password")
	}

	if sa.Username == username && sa.Password == password {
		now := time.Now()
		return Claims{
			Username:  username,
			Admin:     true,
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			Expiry:    now.Add(time.Hour * 24).Unix(),
		}, nil
	}

	return Claims{}, errors.New("invalid username and password")
}
