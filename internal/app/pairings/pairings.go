// Package pairings serves the Pairings web app: a static, client-side page for
// generating and sharing gift-exchange style pairings. The app's files are
// embedded into the binary from the web directory.
package pairings

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed web/index.html web/styles.css web/src/*.js
var files embed.FS

// Set explicitly so browsers always accept the module scripts and stylesheet,
// even on systems whose MIME database maps these extensions oddly.
var contentTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "text/javascript; charset=utf-8",
}

// Handler serves the app. Mount it with its prefix stripped, for example:
//
//	r.Handle("/pairings/*", http.StripPrefix("/pairings", pairings.Handler()))
func Handler() http.Handler {
	web, err := fs.Sub(files, "web")
	if err != nil {
		// Only possible if the embed pattern above is changed incorrectly.
		panic(err)
	}
	fileServer := http.FileServer(http.FS(web))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't list directories such as /src/.
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		if contentType, ok := contentTypes[path.Ext(r.URL.Path)]; ok {
			w.Header().Set("Content-Type", contentType)
		}
		// Embedded files have no modification time, so ask browsers to
		// re-check them; otherwise a redeploy might not be picked up.
		w.Header().Set("Cache-Control", "no-cache")
		fileServer.ServeHTTP(w, r)
	})
}
