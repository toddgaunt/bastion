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
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"toddgaunt.com/monastery/internal/document"
)

type articleVariables struct {
	Title       string
	Description string
	Style       string
	Pinned      map[string]string
	HTML        template.HTML
}

type Content struct {
	Articles map[string]*Article
}

func (content Content) Sorted() []*Article {
	var sorted []*Article
	// Created a list of nested articles sorted by date
	for _, v := range content.Articles {
		sorted = append(sorted, v)
	}

	sort.Slice(sorted, func(i int, j int) bool {
		return sorted[i].Title < sorted[j].Title
	})

	sort.Slice(sorted, func(i int, j int) bool {
		return sorted[i].Created.After(sorted[j].Created)
	})

	return sorted
}

type Article struct {
	Route       string
	Title       string
	Description string
	Author      string
	Created     time.Time
	Updated     time.Time

	HTML    template.HTML
	Problem *ProblemJSON
}

func (article Article) FormattedDate() string {
	return article.Created.Format("2006-01-02")
}

const articlesCtxKey = "articleID"

var articleTemplate = template.Must(template.ParseFiles("templates/article.html"))

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
func GetArticle(content *Content, config Config) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {
		articleID := r.Context().Value(articlesCtxKey).(string)
		article, ok := content.Articles[articleID]

		if !ok {
			return &ProblemJSON{Title: "No Such Article", Status: http.StatusNotFound, Detail: fmt.Sprintf("Article %s does not exist", articleID)}
		}

		if article.Problem != nil {
			return article.Problem
		}

		vars := articleVariables{
			Title:       article.Title,
			Description: article.Description,
			Style:       config.Style,
			Pinned:      config.Pinned,
			HTML:        article.HTML,
		}

		buf := &bytes.Buffer{}
		articleTemplate.Execute(buf, vars)

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	return ProblemHandler(f)
}

// GenerateArticles walks a directory, and generates articles from
// subdirectories and markdown files found
func GenerateArticles(dir string) (map[string]*Article, error) {
	articles := make(map[string]*Article)

	filepath.Walk(dir, func(articlePath string, info os.FileInfo, err error) error {
		articleID := strings.TrimPrefix(articlePath[:len(articlePath)-len(path.Ext(articlePath))], dir+"/")

		if strings.HasPrefix(articleID, ".") {
			// Skip "hidden" files, since '.' is reserved for built-in routes
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			// Don't do anything for directories
			return nil
		}

		article := &Article{Route: articleID}

		data, err := ioutil.ReadFile(articlePath)
		if err != nil {
			article.Problem = &ProblemJSON{Title: "Could not read document file", Status: http.StatusNotFound, Detail: fmt.Sprintf("Content %s could not be read from the filesystem", articleID)}
			articles[articleID] = article
			return nil
		}

		doc, err := document.Parse(data)
		if err != nil {
			article.Problem = &ProblemJSON{Title: "Could not parse document data", Status: http.StatusInternalServerError, Detail: err.Error()}
			articles[articleID] = article
			return nil
		}

		article.Title = doc.Properties.Value("Title")
		article.Description = doc.Properties.Value("Description")
		article.Author = doc.Properties.Value("Author")
		article.Created = ParseDate(doc.Properties.Value("Created"))

		bytes, err := document.GenerateHTML(doc)
		if err != nil {
			article.Problem = &ProblemJSON{Title: "Could not generate HTML from document", Status: http.StatusInternalServerError, Detail: err.Error()}
			articles[articleID] = article
			return nil
		}

		article.HTML = template.HTML(bytes)

		log.Printf("route: %s\n", article.Route)
		articles[articleID] = article

		return nil
	})

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
func ScanContent(config Config) *Content {
	content := &Content{}

	go func() {
		// TODO(todd): Scan periodically
		content.Articles, _ = GenerateArticles(config.ContentPath)
	}()

	return content
}
