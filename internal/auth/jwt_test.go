package auth_test

import (
	"testing"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
)

// testKey was generated with auth.GenerateKey
var testKey = auth.SymmetricKey([32]byte{
	229, 230, 119, 35, 125, 68, 142, 238,
	192, 75, 123, 165, 17, 228, 152, 62,
	79, 230, 115, 170, 37, 211, 237, 192,
	17, 8, 151, 4, 175, 182, 245, 207,
})

func TestSign(t *testing.T) {
	var testcases = []struct {
		name     string
		claims   auth.Claims
		now      time.Time
		duration time.Duration
		key      auth.SymmetricKey

		want auth.JWT
	}{
		{
			name: "Success",
			claims: auth.Claims{
				Username: "samwise",
			},
			now:      time.Unix(0, 0),
			duration: time.Duration(time.Second * 20),
			key:      testKey,

			want: "eyJhbGciOiJIUzI1NiJ9.eyJ1aWQiOiJzYW13aXNlIiwiZXhwIjoyMCwibmJmIjowLCJpYXQiOjB9.mRc0fSnXPD6yAQELEFNoyU4VNg6_GFuYPwpA0rFP42I",
		},
		{
			name: "Failure",
			claims: auth.Claims{
				Username: "smeagol",
				Admin:    true,
			},
			now:      time.Unix(0, 0),
			duration: time.Duration(time.Second * 20),
			key:      testKey,

			want: "eyJhbGciOiJIUzI1NiJ9.eyJ1aWQiOiJzYW13aXNlIiwiZXhwIjoyMCwibmJmIjowLCJpYXQiOjB9.mRc0fSnXPD6yAQELEFNoyU4VNg6_GFuYPwpA0rFP42I",
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.claims.Sign(tc.now, tc.duration, tc.key)

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
		name     string
		token    auth.JWT
		key      auth.SymmetricKey
		now      time.Time
		duration time.Duration

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
			got, err := tc.token.Verify(tc.key)
			if (err == nil) != tc.succeeds {
				t.Fatalf("encryption error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSignAndVerify(t *testing.T) {
	claims := auth.Claims{
		Username:  "samwise.gamgee",
		Admin:     true,
		Expiry:    time.Unix(62, 0).Unix(),
		IssuedAt:  time.Unix(42, 0).Unix(),
		NotBefore: time.Unix(42, 0).Unix(),
	}

	token, err := claims.Sign(time.Unix(42, 0), time.Duration(time.Second*20), testKey)
	if err != nil {
		t.Fatalf("failed to sign valid claims %v: %v", claims, err)
	}

	got, err := token.Verify(testKey)
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
