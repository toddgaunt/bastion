package handlers

import (
	"net/http"

	"github.com/toddgaunt/bastion/internal/errors"
)

var errNotFound = errors.Note{StatusCode: http.StatusNotFound}.Wrap(errors.New("not found"))

func (env Env) NotFound(w http.ResponseWriter, r *http.Request) {
	handleError(w, errNotFound, env.Logger)
}
