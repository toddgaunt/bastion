package handlers_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/handlers"
	"github.com/toddgaunt/bastion/internal/log"
	"github.com/toddgaunt/bastion/internal/tests"
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
		Clock:   tests.MockClock(now),
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
	now := time.Now()
	username := "test username"
	password := "test password"

	authenticator, err := auth.NewSimple(username, password)
	if err != nil {
		t.Fatalf("failed to initialize simple auth: %v", err)
	}

	key, err := auth.GenerateSymmetricKey()
	if err != nil {
		t.Fatalf("failed to initialize signing key: %v", err)
	}

	env := handlers.Env{
		Auth:    authenticator,
		SignKey: key,
		Logger:  log.NewNop(),
		Clock:   tests.MockClock(now),
	}

	testCases := []struct {
		name         string
		setBasicAuth bool
		username     string
		password     string

		wantStatusCode int
		wantHeader     http.Header
	}{
		{
			name:         "ValidCredentials",
			setBasicAuth: true,
			username:     username,
			password:     password,

			wantStatusCode: 200,
			wantHeader: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
		{
			name:         "InvalidPassword",
			setBasicAuth: true,
			username:     username,
			password:     "opensesame",

			wantStatusCode: 401,
			wantHeader: http.Header{
				"Content-Type":     []string{"application/problem+json"},
				"Www-Authenticate": []string{`Basic realm="restricted"`},
			},
		},
		{
			name:         "InvalidUsername",
			setBasicAuth: true,
			username:     username + username,
			password:     password,

			wantStatusCode: 401,
			wantHeader: http.Header{
				"Content-Type":     []string{"application/problem+json"},
				"Www-Authenticate": []string{`Basic realm="restricted"`},
			},
		},
		{
			name:         "NoBasicAuth",
			setBasicAuth: false,

			wantStatusCode: 401,
			wantHeader: http.Header{
				"Content-Type":     []string{"application/problem+json"},
				"Www-Authenticate": []string{`Basic realm="restricted"`},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "http://www.test.com", nil)

			if tc.setBasicAuth {
				r.SetBasicAuth(tc.username, tc.password)
			}

			env.Login(w, r)

			res := w.Result()

			if got, want := res.StatusCode, tc.wantStatusCode; got != want {
				t.Errorf("got status code %d, want %d", got, want)
			}

			if got, want := res.Header, tc.wantHeader; !reflect.DeepEqual(got, want) {
				t.Errorf("got header %v, want %v", got, want)
			}

			got, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			statusCode := res.StatusCode
			contentType := res.Header.Get("Content-Type")
			switch {
			case statusCode == 200 && contentType == "application/json":
				var response struct {
					RefreshToken string `json:"refresh_token"`
					AccessToken  string `json:"access_token"`
				}
				err := json.Unmarshal(got, &response)
				if err != nil {
					t.Fatalf("couldn't unmarshal JSON response: %v", err)
				}

				// These numbers are just for sanity checking, they aren't
				// important other than capturing what is currently expected.
				base64RefreshTokenLength := 24
				savedAccessTokenLength := 179

				if got, want := len(response.RefreshToken), base64RefreshTokenLength; got != want {
					t.Errorf("Got refresh token of size %d, want %d", got, want)
				}
				if got, want := len(response.AccessToken), savedAccessTokenLength; got != want {
					t.Errorf("Got refresh token of size %d, want %d", got, want)
				}
			case statusCode == 401 && contentType == "application/problem+json":
				want := `{"title":"Unauthorized","status":401,"detail":"enter a valid username and password"}`
				if got := string(got); got != want {
					t.Errorf("Document doesn't match what was expected:\n%v",
						string(tests.Diff(want, got)),
					)
				}
			default:
				t.Fatalf("unexpected response code %d and content-type %s", statusCode, contentType)
			}
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
