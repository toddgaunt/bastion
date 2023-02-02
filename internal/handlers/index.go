package handlers

import (
	"bytes"
	_ "embed"
	"net/http"

	"github.com/toddgaunt/bastion/internal/errors"
)

// Index returns an HTTP handler that responds with the site index.
func (env Env) Index(w http.ResponseWriter, r *http.Request) {
	const op = errors.Op("Index")

	// Actions to perform for every request
	fn := func(w http.ResponseWriter, _ *http.Request) errors.Problem {
		details := env.Store.GetDetails()
		vars := templateVariables{
			Title:       details.Name,
			Description: details.Description,
			content:     env.Store,
		}

		buf := &bytes.Buffer{}

		indexTemplate.Execute(buf, vars)

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	err := fn(w, r)
	handleError(w, err, env.Logger)
}
