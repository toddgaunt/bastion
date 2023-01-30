package router

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/toddgaunt/bastion"
)

// New creates a new router for a bastion website.
func New(staticFileServer http.Handler, content bastion.Content, logger bastion.Logger) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.WithValue(logKey, logger))

	r.Route("/", func(r chi.Router) {
		r.Get("/", Handler(GetIndex(indexTemplate, content)))
		r.With(ArticlesCtx).Get("/*", Handler(GetArticle(articleTemplate, content)))
	})

	r.Route("/"+ProblemPath, func(r chi.Router) {
		r.Route("/{problemID}", func(r chi.Router) {
			r.Use(ProblemsCtx)
			r.Get("/", Handler(GetProblem(problemTemplate, content)))
		})
	})

	r.Handle("/.static/*", http.StripPrefix("/.static/", staticFileServer))

	r.Route("/.auth", func(r chi.Router) {
		r.Get("/refresh", Handler(Refresh))
		r.Post("/token", Handler(Token))
	})

	return r, nil
}
