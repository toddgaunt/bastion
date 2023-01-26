package auth_test

import (
	"testing"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/errors"
)

// testKey was generated with auth.GenerateKey
var testKey = auth.SymmetricKey([32]byte{
	229, 230, 119, 35, 125, 68, 142, 238,
	192, 75, 123, 165, 17, 228, 152, 62,
	79, 230, 115, 170, 37, 211, 237, 192,
	17, 8, 151, 4, 175, 182, 245, 207,
})

func TestAuthenticationTokenIsValid(t *testing.T) {
	var testcases = []struct {
		name  string
		token auth.Claims

		want bool
	}{
		{
			name:  "UnexpiredClaims",
			token: auth.NewClaims("Samwise", time.Now(), time.Duration(time.Minute*1000)),

			want: false,
		},
		{
			name:  "ExpiredClaims",
			token: auth.NewClaims("Frodo", time.Now(), time.Duration(time.Minute*0)),

			want: true,
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.token.IsExpired(time.Now()); got != tc.want {
				t.Fatalf("got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEncrypt(t *testing.T) {
	var testcases = []struct {
		name   string
		claims auth.Claims
		key    auth.SymmetricKey

		want auth.JWT
	}{
		{
			name: "Success",
			claims: auth.Claims{
				Username: "Lord Jim",
				Expiry:   time.Unix(20, 0).Unix(),
			},
			key: testKey,

			// Note that this is only the header, since the tail will change each run.
			want: "eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIiwiemlwIjoiREVGIn0..",
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := auth.Encrypt(tc.claims, tc.key)

			if err != nil {
				t.Fatalf("encryption error: %v", err)
			}

			if got := got[:len(tc.want)]; got != tc.want {
				t.Fatalf("got = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	var testcases = []struct {
		name  string
		token auth.JWT
		key   auth.SymmetricKey

		want auth.Claims
		err  error
	}{
		{
			name:  "Success",
			token: "eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIiwiemlwIjoiREVGIn0..AcTuXj2w9QUtz7Jl.PwbJEKBgXD4si-Ok1D9grREeUIGoGJbvYgnFARAFjo6oIEtcfUV1o3lt0JYCYtu8fg.E1DYBKLRTBLb-1rLYBowsg",
			key:   testKey,

			want: auth.Claims{
				Username: "Lord Jim",
				Expiry:   time.Unix(20, 0).Unix(),
			},
			err: nil,
		},
		{
			name:  "WrongAlgorithm",
			token: "eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2IiwiemlwIjoiREVGIn0..NbK-6OUC48LDnCSS5AX8PA.o84W7K3AVZ-AwzXNuh3R_O1h3QhskoZZW8e8jH0wOPfeGI2is2Uh5kUUicc8Q62-wYiWwj3iQ0AA-e97ZVq2tA.vTpJYVf7SSq_4bdlXn5B_w",
			key:   testKey,

			want: auth.Claims{},
			err:  auth.ErrJWTBadAlgo,
		},
		{
			name:  "NotAValidToken",
			token: "not-a-token",
			key:   testKey,

			want: auth.Claims{},
			err:  auth.ErrJWTDecode,
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := auth.Decrypt(tc.token, tc.key)
			if !errors.Is(err, tc.err) {
				t.Fatalf("got err %v, want err %v", err, tc.err)
			}

			if got != tc.want {
				t.Fatalf("got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSign(t *testing.T) {
	var testcases = []struct {
		name   string
		claims auth.Claims
		key    auth.SymmetricKey

		want auth.JWT
	}{
		{
			name: "Success",
			claims: auth.Claims{
				Username: "samwise",
				Expiry:   time.Unix(20, 0).Unix(),
			},
			key: testKey,

			want: "eyJhbGciOiJIUzI1NiJ9.eyJ1aWQiOiJzYW13aXNlIiwiZXhwIjoyMCwibmJmIjowLCJpYXQiOjB9.mRc0fSnXPD6yAQELEFNoyU4VNg6_GFuYPwpA0rFP42I",
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := auth.Sign(tc.claims, tc.key)

			if err != nil {
				t.Fatalf("encryption error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	var testcases = []struct {
		name  string
		token auth.JWT
		key   auth.SymmetricKey

		want     auth.Claims
		succeeds bool
	}{
		{
			name:  "Success",
			token: "eyJhbGciOiJIUzI1NiJ9.eyJ1aWQiOiJzYW13aXNlIiwiZXhwIjoyMCwibmJmIjowLCJpYXQiOjB9.mRc0fSnXPD6yAQELEFNoyU4VNg6_GFuYPwpA0rFP42I",
			key:   testKey,

			want: auth.Claims{
				Username: "samwise",
				Expiry:   time.Unix(20, 0).Unix(),
			},
			succeeds: true,
		},
		{
			name:  "WrongAlgorithm",
			token: "eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2IiwiemlwIjoiREVGIn0..5zlojXBhUEhloH55Qo23ng.oTIdcgK4j6rWvdGTjiS44-2LFplXN-GHw6P5IS1Zvl2B9y9l2FzBJCVeVEpSWsWA7qfgv6P0iud1swyyzTUwiQ.tIe3gVpX72OMofPVe5u40g",
			key:   testKey,

			want:     auth.Claims{},
			succeeds: false,
		},
		{
			name:  "NotAValidToken",
			token: "not-a-token",
			key:   testKey,

			want:     auth.Claims{},
			succeeds: false,
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := auth.Verify(tc.token, tc.key)
			if (err == nil) != tc.succeeds {
				t.Fatalf("encryption error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEncryptAndDecrypt(t *testing.T) {
	claims := auth.Claims{
		Username: "frodo.baggins",
		Expiry:   time.Unix(20, 0).Unix(),
	}

	token, err := auth.Encrypt(claims, testKey)
	if err != nil {
		t.Fatalf("failed to encrypt valid claims %v: %v", claims, err)
	}

	got, err := auth.Decrypt(token, testKey)
	if err != nil {
		t.Fatalf("failed to decrypt token %q: %v", token, err)
	}

	if got != claims {
		t.Fatalf("claims retrieved from token don't match: got %v, want %v", got, claims)
	}
}

func TestSignAndVerify(t *testing.T) {
	claims := auth.Claims{
		Username: "samwise.gamgee",
		Expiry:   time.Unix(20, 0).Unix(),
	}

	token, err := auth.Sign(claims, testKey)
	if err != nil {
		t.Fatalf("failed to sign valid claims %v: %v", claims, err)
	}

	got, err := auth.Verify(token, testKey)
	if err != nil {
		t.Fatalf("failed to verify signature of token %q: %v", token, err)
	}

	if got != claims {
		t.Fatalf("claims retrieved from token don't match: got %v, want %v", got, claims)
	}
}

func TestGenerateSymmetricKey(t *testing.T) {
	t.Parallel()

	// Skip this test for short tests.
	if testing.Short() {
		t.Skip("the function under test reads from /dev/urandom")
	}

	key, err := auth.GenerateSymmetricKey()
	if err != nil {
		t.Fatalf("Key generation failed: %v", err)
	}

	if len(key) != 32 {
		t.Fatalf("Key length not equal to 32 bytes: %v", key)
	}
}
