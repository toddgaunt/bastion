package router

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/errors"
)

// GetIndex returns an HTTP handler that responds with the site index.
func GetIndex(tmpl *template.Template, config content.Config, articleMap *content.ArticleMap) func(w http.ResponseWriter, r *http.Request) error {
	const op = errors.Op("GetIndex")

	// Actions to perform for every request
	return func(w http.ResponseWriter, _ *http.Request) error {
		vars := templateVariables{
			Title:       config.Name,
			Description: config.Description,
			Site:        config,
			ArticleMap:  articleMap,
		}

		buf := &bytes.Buffer{}

		articleMap.Mutex.RLock()
		tmpl.Execute(buf, vars)
		articleMap.Mutex.RUnlock()

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}
}
