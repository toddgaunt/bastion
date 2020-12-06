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
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"toddgaunt.com/monastery/internal/document"
)

type Article struct {
	Route       string
	Title       string
	Description string
	Author      string
	Created     time.Time

	Data    []byte
	Problem *ProblemJSON

	Listing  bool
	Articles map[string]*Article
}

func (article Article) FormattedCreated() string {
	return article.Created.Format("2006-01-02")
}

func (article Article) SortedArticles() []*Article {
	var sorted []*Article
	// Created a list of nested articles sorted by date
	for _, v := range article.Articles {
		sorted = append(sorted, v)
	}
	sort.SliceStable(sorted, func(i int, j int) bool {
		return sorted[i].Created.After(sorted[j].Created)
	})

	return sorted
}

type articleVariables struct {
	Config Config
	Root   *Article
}

const articlesCtxKey = "articleID"

const articleHeaderHTML = siteHeaderHTML
const articleFooterHTML = siteFooterHTML
const articleListingHTML = `<article>
<hr>
{{if .Title}}<h1 id="listing-title">{{.Title}}</h1>
{{end}}{{if .Description}}<p id="listing-description">{{.Description}}</p>
{{end}}<hr>
<ul>{{range $k, $v := .SortedArticles}}
<li><a href="{{$v.Route}}">{{$v.FormattedCreated}} - {{$v.Title}}</a></li>{{end}}
</ul>
</article>
`

var articleHeaderTemplate = template.Must(template.New("articleHeader").Parse(articleHeaderHTML))
var articleFooterTemplate = template.Must(template.New("articleFooter").Parse(articleFooterHTML))
var articleListingTemplate = template.Must(template.New("articleListing").Parse(articleListingHTML))

// ArticlesCtx is middleware for a router to provide a clean path to an article
// for an HTTPHandler
func ArticlesCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := filepath.Clean(chi.URLParam(r, "*"))
		ctx := context.WithValue(r.Context(), articlesCtxKey, articleID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetArticle returns an HTTP handler function to respond to HTTP requests for
// an article. The handler will write an HTML representation of an article as
// a response, or a problemjson response if the article does not exist or there
// was a problem generating it.
func GetArticle(rootDoc *Article, config Config) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {
		articleID := r.Context().Value(articlesCtxKey).(string)

		parts := strings.Split(articleID, "/")

		article := rootDoc
		for _, p := range parts {
			var ok bool
			article, ok = article.Articles[p]
			if !ok {
				return &ProblemJSON{Title: "No Such Article", Status: http.StatusNotFound, Detail: fmt.Sprintf("Article %s does not exist", articleID)}
			}
		}

		if article.Problem != nil {
			return article.Problem
		}

		vars := articleVariables{
			Config: config,
			Root:   rootDoc,
		}

		buf := &bytes.Buffer{}
		articleHeaderTemplate.Execute(buf, vars)
		buf.Write(article.Data)
		articleFooterTemplate.Execute(buf, vars)

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	return ProblemHandler(f)
}

// GenerateArticles walks a directory, and generates articles from
// subdirectories and markdown files found
func GenerateArticles(dir string, reroute func(path string) string) (map[string]*Article, error) {
	files, _ := ioutil.ReadDir(dir)
	articles := make(map[string]*Article)

	for _, f := range files {
		name := f.Name()

		// Skip "hidden" files, since '.' is reserved for built-in routes
		if strings.HasPrefix(name, ".") {
			continue
		}

		articleID := name[:len(name)-len(path.Ext(name))]

		article := &Article{Route: dir + "/" + articleID}
		if reroute != nil {
			article.Route = reroute(article.Route)
		}
		log.Printf("route: %s\n", article.Route)

		if f.IsDir() {
			article.Title = articleID
			var err error
			// Generate a map to nested articles
			if article.Articles, err = GenerateArticles(dir+"/"+name, reroute); err != nil {
				return nil, err
			}

			buf := &bytes.Buffer{}
			articleListingTemplate.Execute(buf, article)
			article.Data = buf.Bytes()
		} else {
			data, err := ioutil.ReadFile(dir + "/" + name)
			if err != nil {
				article.Problem = &ProblemJSON{Title: "Could not read document file", Status: http.StatusNotFound, Detail: fmt.Sprintf("Content %s could not be read from the filesystem", articleID)}
				articles[articleID] = article
				continue
			}

			doc, err := document.Parse(data)
			if err != nil {
				article.Problem = &ProblemJSON{Title: "Could not parse document data", Status: http.StatusInternalServerError, Detail: err.Error()}
				articles[articleID] = article
				continue
			}

			article.Title = doc.Properties.Value("Title")
			article.Description = doc.Properties.Value("Description")
			article.Author = doc.Properties.Value("Author")
			article.Created = ParseDate(doc.Properties.Value("Created"))

			article.Data, err = document.GenerateHTML(doc)
			if err != nil {
				article.Problem = &ProblemJSON{Title: "Could not generate HTML from document", Status: http.StatusInternalServerError, Detail: err.Error()}
				articles[articleID] = article
				continue
			}
		}
		articles[articleID] = article
	}

	return articles, nil
}

// ParseDate parse a string as a date and return a time object, or log
// if the date couldn't be parsed
func ParseDate(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Printf("invalid time: %v", err)
	}

	return t
}

// ScanContent scans for articles for a given configuration
func ScanContent(config Config) *Article {
	rootDoc := &Article{
		Route:       "/",
		Title:       "Root",
		Description: "The root article",
		Problem:     nil,
	}

	go func() {
		// TODO(todd): Scan periodically
		rootDoc.Articles, _ = GenerateArticles(config.ContentPath,
			func(s string) string { return s[len(config.ContentPath):] })
	}()

	return rootDoc
}
