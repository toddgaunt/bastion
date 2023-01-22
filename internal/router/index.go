package router

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/toddgaunt/bastion/internal/articles"
	"github.com/toddgaunt/bastion/internal/httpjson"
)

// GetIndex returns an HTTP handler that responds to requests with the
// Monastery site index
func GetIndex(tmpl *template.Template, config Config, articleMap *articles.ArticleMap) func(w http.ResponseWriter, r *http.Request) {
	// Actions to perform for every request
	f := func(w http.ResponseWriter, _ *http.Request) *httpjson.Problem {
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

	return ProblemHandler(f)
}
