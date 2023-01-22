package auth_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
)

func TestAuthenticationTokenEqual(t *testing.T) {
	var testcases = []struct {
		name string
		this auth.Claims
		that auth.Claims

		want bool
	}{
		{
			"EqualClaims",
			auth.NewClaims("Samwise", time.Duration(time.Minute*5)),
			auth.NewClaims("Samwise", time.Duration(time.Minute*1)),
			true,
		},
		{
			"DifferentClaims",
			auth.NewClaims("Sam", time.Duration(time.Minute*5)),
			auth.NewClaims("Frodo", time.Duration(time.Minute*1)),
			false,
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.this.Equal(tc.that); got != tc.want {
				t.Fatalf("got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAuthenticationTokenIsExpired(t *testing.T) {
	var testcases = []struct {
		name  string
		token auth.Claims

		want bool
	}{
		{
			"UnexpiredClaims",
			auth.NewClaims("Samwise", time.Duration(time.Minute*1000)),

			false,
		},
		{
			"ExpiredClaims",
			auth.NewClaims("Frodo", time.Duration(time.Minute*0)),

			true,
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.token.IsExpired(); got != tc.want {
				t.Fatalf("got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAuthenticationTokenEncrypt(t *testing.T) {
	var testcases = []struct {
		name   string
		claims auth.Claims
		secret []byte

		want auth.JWT
	}{
		{
			"GoodEncryption",
			auth.Claims{
				"sam",
				time.Unix(20, 0).Unix(),
			},
			[]byte("0123456789abcdef"),

			// Note that this is only the header, since the tail will change each run.
			"eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4R0NNIiwiemlwIjoiREVGIn0..",
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := auth.Encrypt(tc.claims, tc.secret)
			fmt.Printf("token: %s\n", got)

			if err != nil {
				t.Fatal("encryption error:", err)
			}

			if got := got[:len(tc.want)]; got != tc.want {
				t.Fatalf("got = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAuthenticationTokenDecrypt(t *testing.T) {
	var testcases = []struct {
		name   string
		token  auth.JWT
		secret []byte

		want     auth.Claims
		succeeds bool
	}{
		{
			"Well encrypted Authentication Token",
			"eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4R0NNIiwiemlwIjoiREVGIn0..k0l-HfuWbJ7s6Uz7.oS9gIDHPT2U3OMXVD2_wYagSLWnw8cUnHA1-dUTQKu2q.ISCidPx38hLzBGFiJ1eRIg",
			[]byte("0123456789abcdef"),

			auth.Claims{
				"Lord Jim",
				time.Unix(20, 0).Unix(),
			},
			true,
		},
		{
			"Wrong algo",
			"eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2IiwiemlwIjoiREVGIn0..5zlojXBhUEhloH55Qo23ng.oTIdcgK4j6rWvdGTjiS44-2LFplXN-GHw6P5IS1Zvl2B9y9l2FzBJCVeVEpSWsWA7qfgv6P0iud1swyyzTUwiQ.tIe3gVpX72OMofPVe5u40g",
			[]byte("0123456789abcdefghijklmnopqrstuv"),

			auth.Claims{},
			false,
		},
		{
			"Poorly encrypted Authentication Token",
			"not-a-token",
			[]byte("0123456789abcdef"),

			auth.Claims{},
			false,
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := auth.Decrypt(tc.token, tc.secret)
			if (err == nil) != tc.succeeds {
				t.Fatal("encryption error:", err)
			}

			if !got.Equal(tc.want) {
				t.Fatalf("got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	t.Parallel()

	// Skip this test for short tests.
	if testing.Short() {
		t.Skip("this test reads from random")
	}

	var key []byte
	var err error
	if key, err = auth.GenerateKey(); err != nil {
		t.Fatal("Key generation failed:", err)
	}

	if got := len(key); got < 16 {
		t.Fatal("Key length less than minimum 16 bytes:", got)
	}
}
