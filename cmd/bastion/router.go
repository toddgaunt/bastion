package main

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/handlers"
	"github.com/toddgaunt/bastion/internal/log"
)

func newRouter(staticFileServer http.Handler, authenticator auth.Authenticator, store content.Store, logger log.Logger) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(log.Middleware(logger))

	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.Index(store))
		r.With(handlers.ArticlePath).Get("/*", handlers.Articles(store))
	})

	r.Route("/.problems", func(r chi.Router) {
		r.Route("/{problemID}", func(r chi.Router) {
			r.Use(handlers.ProblemID)
			r.Get("/", handlers.Problems(store))
		})
	})

	r.Handle("/.static/*", http.StripPrefix("/.static/", staticFileServer))

	r.Route("/.auth", func(r chi.Router) {
		r.Get("/refresh", handlers.Refresh(authenticator))
		r.Post("/token", handlers.Token)
	})

	return r, nil
}
