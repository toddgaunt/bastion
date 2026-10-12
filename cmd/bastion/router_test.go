package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/toddgaunt/bastion/internal/handlers"
)

func TestPairingsRoutes(t *testing.T) {
	r, err := newRouter(http.NotFoundHandler(), handlers.Env{})
	if err != nil {
		t.Fatalf("newRouter: %v", err)
	}

	t.Run("app page", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/.apps/pairings/", nil))
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "<title>Pairings</title>") {
			t.Fatalf("GET /.apps/pairings/: status %d, body %.80q", rec.Code, rec.Body.String())
		}
	})

	t.Run("script", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/.apps/pairings/src/app.js", nil))
		if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/javascript") {
			t.Fatalf("GET /.apps/pairings/src/app.js: status %d, Content-Type %q", rec.Code, rec.Header().Get("Content-Type"))
		}
	})

	t.Run("redirect without trailing slash", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/.apps/pairings", nil))
		if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "/.apps/pairings/" {
			t.Fatalf("GET /.apps/pairings: status %d, Location %q", rec.Code, rec.Header().Get("Location"))
		}
	})
}
