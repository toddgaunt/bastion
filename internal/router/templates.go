package router

import (
	_ "embed"
	"html/template"
	"sort"

	"github.com/toddgaunt/bastion/internal/content"
)

type templateVariables struct {
	Title       string
	Description string
	Site        content.Config
	HTML        template.HTML
	ArticleMap  *content.ArticleMap
}

// Pinned creates a mapping of pinned article titles to their route
func (vars templateVariables) Pinned() map[string]string {
	vars.ArticleMap.Mutex.RLock()
	defer vars.ArticleMap.Mutex.RUnlock()

	var mapping = map[string]string{}
	for _, v := range vars.ArticleMap.Articles {
		// Only add pinned articles to the mapping
		if v.Pinned {
			mapping[v.Title] = v.Route
		}
	}

	return mapping
}

// SortedIndex creates a index of articles sorted by title and created time.
func (vars templateVariables) SortedIndex() []*content.Article {
	vars.ArticleMap.Mutex.RLock()
	defer vars.ArticleMap.Mutex.RUnlock()

	var sorted []*content.Article
	// Created a list of nested articles sorted by date
	for _, v := range vars.ArticleMap.Articles {
		// Only add unpinned articles to the index
		if !v.Pinned {
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

var (
	//go:embed problems.html
	problemTemplateString string
	//go:embed index.html
	indexTemplateString string
	//go:embed articles.html
	articleTemplateString string
)

var (
	indexTemplate   = template.Must(template.New("index").Parse(indexTemplateString))
	articleTemplate = template.Must(template.New("article").Parse(articleTemplateString))
	problemTemplate = template.Must(template.New("problem").Parse(problemTemplateString))
)
