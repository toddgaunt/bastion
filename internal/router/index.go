package router

import (
	"bytes"
	"html/template"
	"net/http"
	"sort"

	"github.com/toddgaunt/bastion/internal/articles"
	"github.com/toddgaunt/bastion/internal/httpjson"
)

const indexTemplateString = `<!DOCTYPE html>
<html>
	<head>
		<title>{{.Title}}</title>
		<meta name="description" content="{{.Description}}">
		<link href="/.static/styles/{{.Site.Style}}.css" type="text/css" rel="stylesheet">
	</head>
	<body>
		<div class="site-navigation">
			<a href="/">{{.Site.Name}}</a>
			{{range $name, $route := .Pinned}}
			<a href="{{$route}}">{{$name}}</a>
			{{end}}
		</div>
		<div class="content">
			<article>
				<div class="article-header">
					<h1 class="article-title">{{.Site.Name}}</h1>
					<p class="article-description">{{.Site.Description}}</p>
				</div>
				<div class="article-body">
					<ul>
						{{range $k, $v := .SortedIndex}}
						<li><a href="{{$v.Route}}">{{$v.FormattedDate}} - {{$v.Title}}</a></li>
						{{end}}
					</ul>
				</div>
			</article>
		</div>
	</body>
</html>`

type indexVariables struct {
	Title       string
	Description string
	Site        Config
	ArticleMap  *articles.ArticleMap
}

// Pinned creates a mapping of pinned article titles to their route
func (vars indexVariables) Pinned() map[string]string {
	vars.ArticleMap.Mutex.RLock()
	defer vars.ArticleMap.Mutex.RUnlock()

	var mapping = map[string]string{}
	for _, v := range vars.ArticleMap.Articles {
		// Only add pinned articles to the mapping
		if v.Pinned == true {
			mapping[v.Title] = v.Route
		}
	}

	return mapping
}

func (vars indexVariables) SortedIndex() []*articles.Article {
	vars.ArticleMap.Mutex.RLock()
	defer vars.ArticleMap.Mutex.RUnlock()

	var sorted []*articles.Article
	// Created a list of nested articles sorted by date
	for _, v := range vars.ArticleMap.Articles {
		// Only add unpinned articles to the index
		if v.Pinned == false {
			sorted = append(sorted, v)
		}
	}

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
func GetIndex(tmpl *template.Template, config Config, articleMap *articles.ArticleMap) func(w http.ResponseWriter, r *http.Request) {
	// Actions to perform for every request
	f := func(w http.ResponseWriter, r *http.Request) *httpjson.Problem {
		vars := indexVariables{
			Title:       config.Name,
			Description: config.Description,
			Site:        config,
			ArticleMap:  articleMap,
		}

		buf := &bytes.Buffer{}

		articleMap.Mutex.RLock()
		tmpl.Execute(buf, vars)
		articleMap.Mutex.RUnlock()

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	return ProblemHandler(f)
}
