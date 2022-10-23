package router

import (
	"html/template"
	"net/http"
	"path"
	"sync"

	"github.com/go-chi/chi"
	"github.com/toddgaunt/bastion/internal/articles"
)

// Config contains all configuration for a bastion server's router.
type Config struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Style        string            `json:"style"`
	Pinned       map[string]string `json:"pinned"`
	ScanInterval int               `json:"scan_interval"`
}

var (
	indexTemplate   = template.Must(template.New("index").Parse(indexTemplateString))
	articleTemplate = template.Must(template.New("article").Parse(articleTemplateString))
	problemTemplate = template.Must(template.New("problem").Parse(problemTemplateString))
)

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
