package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/errors"
)

const problemsCtxKey = contextKey("problemID")

// ProblemID extracts the problem ID from the request URL
func ProblemID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		problemID := chi.URLParam(r, "problemID")
		ctx := context.WithValue(r.Context(), problemsCtxKey, problemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Problems is a request handler that returns an HTTP handler that responds
// to a request with a document describing a particular problem.
func (env Env) Problems(w http.ResponseWriter, r *http.Request) {
	const op = "Problems"

	fn := func(w http.ResponseWriter, r *http.Request) errors.Problem {
		problemID := r.Context().Value(problemsCtxKey).(string)

		description := ""

		switch problemID {
		case "article-not-found":
			description = `This article does not exist`
		case "not-found":
			description = `There was no content available`
		case "internal-server-error":
			description = `The server experienced an error which was no fault of the client`
		default:
			return errors.Note{
				Op:         op,
				StatusCode: http.StatusNotFound,
				Detail:     fmt.Sprintf("Documenation for %s is not available", problemID),
			}.Wrap(errors.New("problem not registered"))
		}

		vars := templateVariables{
			Title:       problemID,
			Description: description,
			content:     env.Store,
		}

		buf := &bytes.Buffer{}
		problemTemplate.Execute(buf, vars)
		w.Write(buf.Bytes())

		return nil
	}

	err := fn(w, r)
	handleError(w, err, env.Logger)

}
