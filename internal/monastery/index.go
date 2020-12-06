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
	"html/template"
	"net/http"
)

type indexVariables struct {
	Title       string
	Description string
	Style       string
	Pinned      map[string]string
	Content     *Content
}

var indexTemplate = template.Must(template.ParseFiles("templates/index.html"))

// GetIndex returns an HTTP handler that responds to requests with the
// Monastery site index
func GetIndex(content *Content, config Config) func(w http.ResponseWriter, r *http.Request) {
	// Actions to perform for every request
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {
		vars := indexVariables{
			Title:       config.Title,
			Description: config.Description,
			Style:       config.Style,
			Pinned:      config.Pinned,
			Content:     content,
		}

		buf := &bytes.Buffer{}
		indexTemplate.Execute(buf, vars)
		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	return ProblemHandler(f)
}
