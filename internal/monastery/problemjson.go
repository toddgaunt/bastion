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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const ProblemPath = ".problems"

// ProblemJSON for representing RFC 7807
type ProblemJSON struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// ProblemHandler wraps an HTTP handler that returns a ProblemJSON as a standard
// HTTP handler that returns nothing. This allows for proper error propogation
// while being compatible with the standard HTTP handler API
func ProblemHandler(
	f func(w http.ResponseWriter, r *http.Request) *ProblemJSON,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		problem := f(w, r)
		if problem == nil {
			//w.WriteHeader(http.StatusOK)
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
			protocol := "https"

			urlTitle := problem.Title
			urlTitle = strings.ReplaceAll(urlTitle, " ", "-")
			urlTitle = strings.ToLower(urlTitle)
			urlTitle = url.PathEscape(urlTitle)

			if r.TLS == nil {
				protocol = "http"
			}
			problem.Type = fmt.Sprintf(
				"%s://%s/%s/%s",
				protocol,
				r.Host,
				ProblemPath,
				urlTitle,
			)
		}

		bytes, _ := json.MarshalIndent(problem, "", "  ")
		w.Header().Add("Content-Type", "application/problem+json")
		w.WriteHeader(problem.Status)
		w.Write(bytes)
	}
}
