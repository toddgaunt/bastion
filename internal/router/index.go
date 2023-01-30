package router

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/toddgaunt/bastion"
	"github.com/toddgaunt/bastion/internal/errors"
)

// GetIndex returns an HTTP handler that responds with the site index.
func GetIndex(tmpl *template.Template, content bastion.Content) func(w http.ResponseWriter, r *http.Request) error {
	const op = errors.Op("GetIndex")

	// Actions to perform for every request
	return func(w http.ResponseWriter, _ *http.Request) error {
		details := content.Details()
		vars := templateVariables{
			Title:       details.Name,
			Description: details.Description,
			content:     content,
		}

		buf := &bytes.Buffer{}

		tmpl.Execute(buf, vars)

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}
}
