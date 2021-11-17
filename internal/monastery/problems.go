// Copyright 2020, Todd Gaunt <toddgaunt@protonmail.com>
//
// This file is part of Monastery.
//
// Monastery is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Monastery is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Monastery.  If not, see <https://www.gnu.org/licenses/>.

package monastery

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
	Site        SiteConfig
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
func GetProblem(tmpl *template.Template, config SiteConfig) func(w http.ResponseWriter, r *http.Request) {
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
