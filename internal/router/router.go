package router

import (
	_ "embed"
	"net/http"
	"path"
	"sync"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/articles"
)

type contextKey string

// Config contains all configuration for a bastion server's router.
type Config struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Style        string `json:"style"`
	ScanInterval int    `json:"scan_interval"`
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
			r.Get("/", GetProblem(problemTemplate, config, articles))
		})
	})

	r.Handle("/.static/*", http.StripPrefix("/.static/", staticFileServer))

	// Closing this channel signals all worker threads to stop and cleanup.
	//close(done)
	//wg.Wait()

	return r, nil
}
