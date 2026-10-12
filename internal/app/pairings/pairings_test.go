package pairings_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/toddgaunt/bastion/internal/app/pairings"
)

func TestHandler(t *testing.T) {
	handler := pairings.Handler()

	tests := []struct {
		path        string
		status      int
		contentType string
		contains    string
	}{
		{"/", http.StatusOK, "text/html", "<title>Pairings</title>"},
		{"/styles.css", http.StatusOK, "text/css", "--bg"},
		{"/src/app.js", http.StatusOK, "text/javascript", "from './core.js'"},
		{"/src/core.js", http.StatusOK, "text/javascript", "export function decodeCode"},
		{"/src/", http.StatusNotFound, "", ""},
		{"/missing.js", http.StatusNotFound, "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, tc.contentType) {
				t.Errorf("Content-Type = %q, want prefix %q", got, tc.contentType)
			}
			if !strings.Contains(rec.Body.String(), tc.contains) {
				t.Errorf("body does not contain %q", tc.contains)
			}
		})
	}
}
