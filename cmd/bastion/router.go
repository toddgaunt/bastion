package main

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/toddgaunt/bastion/internal/handlers"
)

func newRouter(staticFileServer http.Handler, env handlers.Env) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.NotFound(env.NotFound)

	r.Route("/", func(r chi.Router) {
		r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/.static/favicon.ico", http.StatusMovedPermanently)
		})
		r.Get("/", env.Index)
		r.Group(func(r chi.Router) {
			r.Use(handlers.ArticlePath)
			r.Get("/*", env.GetArticle)
			r.With(env.Authorize).Post("/*", env.UpdateDocument)
		})
	})

	r.Route("/.problems", func(r chi.Router) {
		r.Route("/{problemID}", func(r chi.Router) {
			r.Use(handlers.ProblemID)
			r.Get("/", env.Problems)
		})
	})

	r.Handle("/.static/*", http.StripPrefix("/.static/", staticFileServer))

	r.Route("/.auth", func(r chi.Router) {
		r.Get("/login", env.Login)
		r.Post("/token", env.Token)
	})

	return r, nil
}
