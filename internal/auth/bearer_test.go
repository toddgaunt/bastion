package auth_test

import (
	"net/http"
	"testing"

	"github.com/toddgaunt/bastion/internal/auth"
)

func TestBearerFromHeader(t *testing.T) {
	t.Parallel()

	var testcases = []struct {
		name string
		in   http.Header

		want auth.JWT
	}{
		{
			"GoodHeader",
			http.Header{"Authorization": {"Bearer foobar"}},

			"foobar",
		},
		{
			"BadBearerHeaderMispelling",
			http.Header{"Authorization": {"Boorer foobar"}},

			"",
		},
		{
			"BadBearerHeaderTooLong",
			http.Header{"Authorization": {"Bearers foobar"}},

			"",
		},
		{
			"BadBearerHeaderTooShort",
			http.Header{"Authorization": {"B foobar"}},

			"",
		},
		{
			"Noheader",
			nil,

			"",
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := auth.GetTokenFromBearer(tc.in); got != tc.want {
				t.Fatalf("got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAddBearerToHeader(t *testing.T) {
	t.Parallel()

	var testcases = []struct {
		name string
		in   auth.JWT

		want string
	}{
		{
			"AddBearerHeader",
			"foobar",

			"Bearer foobar",
		},
	}

	for _, tc := range testcases {
		var tc = tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var header = make(http.Header)
			auth.AddTokenAsBearer(header, tc.in)
			if got := header.Get("Authorization"); got != tc.want {
				t.Fatalf("got = %v, want %v", got, tc.want)
			}
		})
	}
}
