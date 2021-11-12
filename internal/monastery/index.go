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
	"sort"
)

type indexVariables struct {
	Title       string
	Description string
	Style       string
	Pinned      map[string]string
	Content     *Content
}

var indexTemplate = template.Must(template.ParseFiles("templates/index.html"))

func (vars indexVariables) SortedIndex() []*Article {
	var sorted []*Article

	//NOTE: critical section begin
	vars.Content.mutex.RLock()
	// Created a list of nested articles sorted by date
	for _, v := range vars.Content.Articles {
		// Only add unpinned articles to the index
		if _, ok := vars.Pinned[v.Title]; !ok {
			sorted = append(sorted, v)
		}
	}
	vars.Content.mutex.RUnlock()
	//NOTE: critical section end

	sort.Slice(sorted, func(i int, j int) bool {
		return sorted[i].Title < sorted[j].Title
	})

	sort.Slice(sorted, func(i int, j int) bool {
		return sorted[i].Created.After(sorted[j].Created)
	})

	return sorted
}

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

		content.mutex.RLock()
		indexTemplate.Execute(buf, vars)
		content.mutex.RUnlock()

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	return ProblemHandler(f)
}
