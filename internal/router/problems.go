package router

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/errors"
)

const ProblemPath = ".problems"

const problemsCtxKey = contextKey("problemID")

// ProblemsCtx is a middleware that extracts the problem ID
// from the URL of the HTTP request.
func ProblemsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		problemID := chi.URLParam(r, "problemID")
		ctx := context.WithValue(r.Context(), problemsCtxKey, problemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetProblem is a request handler that returns an HTTP handler that responds
// to a request with a document describing a particular problem.
func GetProblem(tmpl *template.Template, config content.Config, articles *content.ArticleMap) func(w http.ResponseWriter, r *http.Request) error {
	const op = "GetProblem"

	return func(w http.ResponseWriter, r *http.Request) error {
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
			return errors.Annotation{
				WithOp:     op,
				WithStatus: http.StatusNotFound,
				WithDetail: fmt.Sprintf("Documenation for %s is not available", problemID),
			}.Wrap(errors.New("problem not registered"))
		}

		vars := templateVariables{
			Title:       problemID,
			Description: description,
			Site:        config,
			ArticleMap:  articles,
		}

		buf := &bytes.Buffer{}
		tmpl.Execute(buf, vars)
		w.Write(buf.Bytes())

		return nil
	}
}
