package router

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/log"
)

type contextKey string

// New creates a new router for a bastion website.
func New(staticFileServer http.Handler, articles *content.ArticleMap, config content.Config) (chi.Router, error) {
	r := chi.NewRouter()
	logger := log.New()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(log.Middleware(logger))

	r.Route("/", func(r chi.Router) {
		r.Get("/", Handler(GetIndex(indexTemplate, config, articles)))
		r.With(ArticlesCtx).Get("/*", Handler(GetArticle(articleTemplate, config, articles)))
	})

	r.Route("/"+ProblemPath, func(r chi.Router) {
		r.Route("/{problemID}", func(r chi.Router) {
			r.Use(ProblemsCtx)
			r.Get("/", Handler(GetProblem(problemTemplate, config, articles)))
		})
	})

	r.Handle("/.static/*", http.StripPrefix("/.static/", staticFileServer))

	r.Route("/.auth", func(r chi.Router) {
		r.Get("/refresh", Handler(Refresh))
		r.Post("/token", Handler(Token))
	})

	return r, nil
}
