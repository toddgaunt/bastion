package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/handlers"
	"github.com/toddgaunt/bastion/internal/log"
)

var noopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func mustSign(
	t *testing.T,
	key auth.SymmetricKey,
	claims auth.Claims,
	now time.Time,
	lifetime time.Duration,
) string {
	jwt, err := key.Sign(claims, now, lifetime)
	if err != nil {
		t.Fatalf("failed to sign claims: %v", err)
	}
	return string(jwt)
}

func TestAuthorize(t *testing.T) {
	now := time.Now()

	key, err := auth.GenerateSymmetricKey()
	if err != nil {
		t.Fatalf("failed to initialize signing key: %v", err)
	}

	env := handlers.Env{
		SignKey: key,
		Logger:  log.NewNop(),
	}

	testCases := []struct {
		name  string
		token string

		wantAuthorized bool
		wantHeader     http.Header
		wantBody       []byte
	}{
		{
			name:  "UnexpiredToken",
			token: mustSign(t, key, auth.Claims{}, now, time.Second),

			wantAuthorized: true,
			wantHeader:     nil,
			wantBody:       nil,
		},
		{
			name:  "TokenNotYetValid",
			token: mustSign(t, key, auth.Claims{}, now.Add(time.Minute), time.Minute),

			wantAuthorized: false,
			wantHeader:     nil,
			wantBody:       nil,
		},
		{
			name:  "ExpiredToken",
			token: mustSign(t, key, auth.Claims{}, now.Add(-time.Second), 0),

			wantAuthorized: false,
			wantHeader:     nil,
			wantBody:       nil,
		},
		{
			name:  "InvalidToken",
			token: "not-a-valid-jwt",

			wantAuthorized: false,
			wantHeader:     nil,
			wantBody:       nil,
		},
		{
			name:  "EmptyHeader",
			token: "",

			wantAuthorized: false,
			wantHeader:     nil,
			wantBody:       nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			authorized := false
			testHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				authorized = true
			})

			handler := env.Authorize(testHandler)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "http://www.test.com", nil)
			r.Header.Add("Authorization", tc.token)

			handler.ServeHTTP(w, r)

			if authorized == true && tc.wantAuthorized == false {
				t.Fatalf("authorized request when it should have been rejected")
			}

			if authorized == false && tc.wantAuthorized == true {
				t.Fatalf("failed to authorized request when it should have been")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "test name",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}

func TestToken(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "test name",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
