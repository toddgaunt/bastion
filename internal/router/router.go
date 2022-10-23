package router

import (
	"html/template"
	"net/http"
	"path"
	"sort"
	"sync"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/articles"
)

// Config contains all configuration for a bastion server's router.
type Config struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Style        string `json:"style"`
	ScanInterval int    `json:"scan_interval"`
}

var (
	indexTemplate   = template.Must(template.New("index").Parse(indexTemplateString))
	articleTemplate = template.Must(template.New("article").Parse(articleTemplateString))
	problemTemplate = template.Must(template.New("problem").Parse(problemTemplateString))
)

type templateVariables struct {
	Title       string
	Description string
	Site        Config
	HTML        template.HTML
	ArticleMap  *articles.ArticleMap
}

// Pinned creates a mapping of pinned article titles to their route
func (vars templateVariables) Pinned() map[string]string {
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

func (vars templateVariables) SortedIndex() []*articles.Article {
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

// New creates a new router for a bastion website.
func New(prefixDir string, config Config) (chi.Router, error) {
	dir := path.Clean(prefixDir)
	r := chi.NewRouter()

	var done chan bool
	var wg = &sync.WaitGroup{}

	staticFileServer := http.FileServer(http.Dir(dir + "/static"))
	articles := articles.IntervalScan(dir+"/articles", config.ScanInterval, done, wg)

	r.Route("/", func(r chi.Router) {
		r.Get("/", GetIndex(indexTemplate, config, articles))
		r.With(ArticlesCtx).Get("/*", GetArticle(articleTemplate, config, articles))
	})

	r.Route("/"+ProblemPath, func(r chi.Router) {
		r.Route("/{problemID}", func(r chi.Router) {
			r.Use(ProblemsCtx)
			r.Get("/", GetProblem(problemTemplate, config))
		})
	})

	r.Handle("/.static/*", http.StripPrefix("/.static/", staticFileServer))

	// Closing this channel signals all worker threads to stop and cleanup.
	//close(done)
	//wg.Wait()

	return r, nil
}
