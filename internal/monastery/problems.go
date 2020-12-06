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
	Config Config
	Root   *Article
	Title  string
	Detail string
}

const problemsCtxKey = "problemID"

const problemHeaderHTML = siteHeaderHTML + `
<article>
<hr>
<h1 id="problem-title">{{.Title}}</h1>
<hr>
`

const problemFooterHTML = siteFooterHTML + `
</article>
</body>
</html>`

var problemHeaderTemplate = template.Must(template.New("problemHeader").Parse(problemHeaderHTML))
var problemFooterTemplate = template.Must(template.New("problemFooter").Parse(problemFooterHTML))

func ProblemsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		problemID := chi.URLParam(r, problemsCtxKey)
		ctx := context.WithValue(r.Context(), problemsCtxKey, problemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetProblem returns an HTTP handler that responds to a request with a document
// describing a particular problem
func GetProblem(rootDoc *Article, config Config) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {
		problemID := r.Context().Value(problemsCtxKey).(string)

		description := ""

		switch problemID {
		case "not-found":
			description = `<p>There was no content available</p>`
		case "internal-server-error":
			description = `<p>The server experienced an error which was no fault of the client</p>`
		default:
			return &ProblemJSON{Status: http.StatusNotFound, Detail: fmt.Sprintf("Explanation for %s does not exist", problemID)}
		}

		vars := problemVariables{
			Config: config,
			Root:   rootDoc,
			Title:  problemID,
			Detail: problemID,
		}

		buf := &bytes.Buffer{}
		problemHeaderTemplate.Execute(buf, vars)
		buf.WriteString(description)
		problemFooterTemplate.Execute(buf, vars)
		w.Write([]byte(buf.String()))

		return nil
	}

	return ProblemHandler(f)
}
