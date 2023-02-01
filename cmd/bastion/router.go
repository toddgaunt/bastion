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

	r.Route("/", func(r chi.Router) {
		r.Get("/", env.Index)
		r.With(handlers.ArticlePath).Get("/*", env.Articles)
	})

	r.Route("/.problems", func(r chi.Router) {
		r.Route("/{problemID}", func(r chi.Router) {
			r.Use(handlers.ProblemID)
			r.Get("/", env.Problems)
		})
	})

	r.Handle("/.static/*", http.StripPrefix("/.static/", staticFileServer))

	r.Route("/.auth", func(r chi.Router) {
		r.Get("/login", env.Refresh)
		r.Post("/refresh", env.Token)
	})

	return r, nil
}
