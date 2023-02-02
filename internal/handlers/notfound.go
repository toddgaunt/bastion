package handlers

import (
	"net/http"

	"github.com/toddgaunt/bastion/internal/errors"
)

var errNotFound = errors.Annotation{WithStatus: http.StatusBadRequest}.Wrap(errors.New("not found"))

func (env Env) NotFound(w http.ResponseWriter, r *http.Request) {
	handleError(w, errNotFound, env.Logger)
}
