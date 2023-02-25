package auth_test

import (
	"strings"
	"testing"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

func TestSimple(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string

		err error
	}{
		{
			name:     "ValidUsernameAndPassword",
			username: "name",
			password: "password",
		},
		{
			name:     "EmptyUsername",
			username: "",
			password: "password",

			err: auth.SimpleInitErr,
		},
		{
			name:     "EmptyPassword",
			username: "name",
			password: "",

			err: auth.SimpleInitErr,
		},
		{
			name:     "EmptyUsernameAndPassword",
			username: "",
			password: "",

			err: auth.SimpleInitErr,
		},
		{
			name:     "LongPassword",
			username: "",
			password: strings.Repeat("12345678", 9),

			err: auth.SimpleInitErr,
		},
		{
			name:     "TooLongPassword",
			username: "name",
			// bcrypt doesn't accept passwords over 72 bytes long.
			password: strings.Repeat("12345678", 9) + "!",

			err: bcrypt.ErrPasswordTooLong,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := auth.NewSimple(tc.username, tc.password)
			if !errors.Is(err, tc.err) {
				t.Fatalf("got err %v, want %v", err, tc.err)
			}

			if err != nil {
				if got != nil {
					t.Fatalf("got nil when something was expected")
				}
			}

			if err == nil {
				if got == nil {
					t.Fatalf("got something when nil was expected")
				}
			}
		})
	}
}
