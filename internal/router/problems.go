package router

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
)

type problemVariables struct {
	Title       string
	Description string
	Site        Config
}

const problemsCtxKey = "problemID"

func ProblemsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		problemID := chi.URLParam(r, problemsCtxKey)
		ctx := context.WithValue(r.Context(), problemsCtxKey, problemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetProblem returns an HTTP handler that responds to a request with a document
// describing a particular problem
func GetProblem(tmpl *template.Template, config Config) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {
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
			return &ProblemJSON{Status: http.StatusNotFound, Detail: fmt.Sprintf("Explanation for %s does not exist", problemID)}
		}

		vars := problemVariables{
			Title:       problemID,   // problem.Title
			Description: description, // problem.Detail
			Site:        config,
		}

		buf := &bytes.Buffer{}
		tmpl.Execute(buf, vars)
		w.Write([]byte(buf.String()))

		return nil
	}

	return ProblemHandler(f)
}
