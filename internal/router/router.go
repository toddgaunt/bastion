package router

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"bastionburrow.com/bastion/internal/content"
	"github.com/go-chi/chi"
)

func New(prefixDir string, config Config) (chi.Router, error) {
	r := chi.NewRouter()

	indexTemplate, err := template.ParseFiles(prefixDir + "/templates/index.html")
	if err != nil {
		return nil, fmt.Errorf("couldn't load index template: %w", err)
	}
	articleTemplate, err := template.ParseFiles(prefixDir + "/templates/article.html")
	if err != nil {
		return nil, fmt.Errorf("couldn't load article template: %w", err)
	}
	problemTemplate, err := template.ParseFiles(prefixDir + "/templates/problem.html")
	if err != nil {
		return nil, fmt.Errorf("couldn't load problem template: %w", err)
	}

	var done chan bool
	var wg sync.WaitGroup

	staticFileServer := http.FileServer(http.Dir(prefixDir + "/static"))
	content := content.IntervalScan(prefixDir+"/content", config.ScanInterval, done, wg)

	r.Route("/", func(r chi.Router) {
		r.Get("/", GetIndex(indexTemplate, config, content))
		r.With(ArticlesCtx).Get("/*", GetArticle(articleTemplate, config, content))
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
