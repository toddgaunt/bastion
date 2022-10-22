package router

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/httpjson"
)

const ProblemPath = ".problems"

// ProblemHandler wraps an HTTP handler that returns a httpjson.Problem as a standard
// HTTP handler that returns nothing. This allows for proper error propogation
// while being compatible with the standard HTTP handler API
func ProblemHandler(
	handlerFunc func(w http.ResponseWriter, r *http.Request) *httpjson.Problem,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var problem = handlerFunc(w, r)
		if problem == nil {
			return
		}

		// Fill in values that we left unfilled
		if problem.Status == 0 {
			problem.Status = http.StatusInternalServerError
		}

		if problem.Title == "" {
			problem.Title = http.StatusText(problem.Status)
		}

		if problem.Type == "" {
			urlTitle := problem.Title
			urlTitle = strings.ReplaceAll(urlTitle, " ", "-")
			urlTitle = strings.ToLower(urlTitle)
			urlTitle = url.PathEscape(urlTitle)

			var scheme = "https"
			if r.TLS == nil {
				scheme = "http"
			}

			problem.Type = fmt.Sprintf(
				"%s://%s/%s/%s",
				scheme,
				r.Host,
				ProblemPath,
				urlTitle,
			)
		}

		httpjson.WriteProblem(w, *problem)
	}
}

type problemVariables struct {
	Title       string
	Description string
	Site        Config
}

const problemsCtxKey = "problemID"

// ProblemsCtx is a middleware that extracts the problem ID
// from the URL of the HTTP request.
func ProblemsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		problemID := chi.URLParam(r, problemsCtxKey)
		ctx := context.WithValue(r.Context(), problemsCtxKey, problemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetProblem is a request handler that returns an HTTP handler that responds
// to a request with a document describing a particular problem.
func GetProblem(tmpl *template.Template, config Config) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) *httpjson.Problem {
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
			return &httpjson.Problem{Status: http.StatusNotFound, Detail: fmt.Sprintf("Explanation for %s does not exist", problemID)}
		}

		vars := problemVariables{
			Title:       problemID,
			Description: description,
			Site:        config,
		}

		buf := &bytes.Buffer{}
		tmpl.Execute(buf, vars)
		w.Write([]byte(buf.String()))

		return nil
	}

	return ProblemHandler(f)
}
