package router

import (
	"bytes"
	"html/template"
	"net/http"
	"sort"
)

type indexVariables struct {
	Title       string
	Description string
	Site        Config
	Content     *Content
}

func (vars indexVariables) SortedIndex() []*Article {
	var sorted []*Article

	//NOTE: critical section begin
	vars.Content.mutex.RLock()
	// Created a list of nested articles sorted by date
	for _, v := range vars.Content.Articles {
		// Only add unpinned articles to the index
		if _, ok := vars.Site.Pinned[v.Title]; !ok {
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
func GetIndex(tmpl *template.Template, config Config, content *Content) func(w http.ResponseWriter, r *http.Request) {
	// Actions to perform for every request
	f := func(w http.ResponseWriter, r *http.Request) *ProblemJSON {
		vars := indexVariables{
			Title:       config.Name,
			Description: config.Description,
			Site:        config,
			Content:     content,
		}

		buf := &bytes.Buffer{}

		content.mutex.RLock()
		tmpl.Execute(buf, vars)
		content.mutex.RUnlock()

		w.Header().Add("Content-Type", "text/html")
		w.Write(buf.Bytes())

		return nil
	}

	return ProblemHandler(f)
}