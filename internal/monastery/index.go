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
	Config Config
	Root   *Article
	Latest []*Article
}

const indexHTML = siteHeaderHTML +
	`<article>
	<hr>
	<h1 id="listing-title">Latest</h1>
	<hr>
	<ul>{{range $k, $v := .Latest}}
	<li><a href="{{$v.Route}}">{{$v.FormattedCreated}} - {{$v.Title}}</a></li>{{end}}
	</ul>
	</article>` +
	siteFooterHTML

var indexTemplate = template.Must(template.New("index").Parse(indexHTML))

func flattenArticles(rootDoc *Article, acc []*Article) []*Article {
	for _, v := range rootDoc.Articles {
		if v.Articles != nil {
			acc = append(acc, flattenArticles(v, acc)...)
		} else {
			acc = append(acc, v)
		}
	}

	return acc
}

func latestArticles(rootDoc *Article) []*Article {
	latest := flattenArticles(rootDoc, []*Article{})
	sort.SliceStable(latest, func(i int, j int) bool {
		return latest[i].Created.After(latest[j].Created)
	})

	if len(latest) > 5 {
		return latest[:5]
	} else {
		return latest
	}
}

// GetIndex returns an HTTP handler that responds to requests with the
// Monastery site index
func GetIndex(rootDoc *Article, config Config) func(w http.ResponseWriter, r *http.Request) {
	// Actions to perform for every request
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {
		vars := indexVariables{
			Config: config,
			Root:   rootDoc,
			Latest: latestArticles(rootDoc),
		}

		buf := &bytes.Buffer{}
		indexTemplate.Execute(buf, vars)
		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	return ProblemHandler(f)
}
