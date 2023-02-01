package handlers

import (
	"bytes"
	_ "embed"
	"net/http"

	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/errors"
)

// Index returns an HTTP handler that responds with the site index.
func Index(store content.Store) func(w http.ResponseWriter, r *http.Request) {
	const op = errors.Op("GetIndex")

	// Actions to perform for every request
	fn := func(w http.ResponseWriter, _ *http.Request) error {
		details := store.GetDetails()
		vars := templateVariables{
			Title:       details.Name,
			Description: details.Description,
			content:     store,
		}

		buf := &bytes.Buffer{}

		indexTemplate.Execute(buf, vars)

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	return wrapper(fn)
}
