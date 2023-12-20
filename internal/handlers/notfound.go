package handlers

import (
	"net/http"

	"github.com/toddgaunt/bastion/internal/errors"
)

var errNotFound = errors.Note{StatusCode: http.StatusNotFound}.Wrap(errors.New("not found"))

// NotFound is the default handler for pages that do not exist.
func (env Env) NotFound(w http.ResponseWriter, r *http.Request) {
	handleError(w, errNotFound, env.Logger)
}
