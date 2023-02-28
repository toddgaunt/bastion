package auth_test

import (
	"testing"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
)

func TestClaimsIsValid(t *testing.T) {
	var now = time.Now()

	var testcases = []struct {
		name  string
		token auth.Claims
		want  bool
	}{
		{
			"ValidClaims",
			auth.Claims{
				IssuedAt:  now.Unix(),
				NotBefore: now.Unix(),
				Expiry:    now.Add(time.Minute).Unix(),
			},
			true,
		},
		{
			"ExpiredClaims",
			auth.Claims{
				IssuedAt:  now.Unix(),
				NotBefore: now.Unix(),
				Expiry:    now.Unix(),
			},
			false,
		},
		{
			"NotBeforeInvalidClaims",
			auth.Claims{
				IssuedAt:  now.Unix(),
				NotBefore: now.Add(time.Second).Unix(),
				Expiry:    now.Add(time.Minute).Unix(),
			},
			false,
		},
		{
			"NotBeforeInvalidAndExpiredClaims",
			auth.Claims{
				IssuedAt:  now.Unix(),
				NotBefore: now.Add(time.Second).Unix(),
				Expiry:    now.Unix(),
			},
			false,
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.token.IsValid(now); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
