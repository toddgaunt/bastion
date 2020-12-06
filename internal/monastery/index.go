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
	Config Config
	Root   *Article
}

const indexHTML = siteHeaderHTML +
	"<article><p>{{if .Config.Description}}{{.Config.Description}}{{else}}No description available{{end}}</p></article>" +
	siteFooterHTML

var indexTemplate = template.Must(template.New("index").Parse(indexHTML))

// GetIndex returns an HTTP handler that responds to requests with the
// Monastery site index
func GetIndex(rootDoc *Article, config Config) func(w http.ResponseWriter, r *http.Request) {
	vars := indexVariables{
		Config: config,
		Root:   rootDoc,
	}

	// Actions to perform for every request
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {

		buf := &bytes.Buffer{}
		indexTemplate.Execute(buf, vars)
		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())
		return nil
	}

	return ProblemHandler(f)
}
