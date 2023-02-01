package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/log"
)

// Env stores all application state that isn't request specific. All fields
// must be safe for concurrent use.
type Env struct {
	Auth   auth.Authenticator
	Store  content.Store
	Logger log.Logger
}

// Problem represents server errors in JSON, defined by IETF RFC 7807.
type Problem struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// Wrap wraps an HTTP handler that returns a httpjson.Problem so it
// can be logged and written to a user after a handler returns it
func (e Env) Wrap(handlerFunc func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handlerFunc(w, r)
		if err == nil {
			return
		}

		logger := e.Logger

		if err, ok := err.(interface{ Fields() map[string]any }); ok {
			for k, v := range err.Fields() {
				logger = logger.With(k, v)
			}
		}

		problem := Problem{}

		if err, ok := err.(interface{ Type() string }); ok {
			problem.Type = err.Type()
		}

		if err, ok := err.(interface{ Title() string }); ok {
			problem.Title = err.Title()
		}

		if err, ok := err.(interface{ Status() int }); ok {
			problem.Status = err.Status()
		}

		if err, ok := err.(interface{ Detail() string }); ok {
			problem.Detail = err.Detail()
		}

		// Fill in default values.

		if problem.Status == 0 {
			problem.Status = http.StatusInternalServerError
		}

		if problem.Title == "" {
			problem.Title = http.StatusText(problem.Status)
		}

		// Log values as structured fields.

		logger = logger.With("status", problem.Status)

		logger = logger.With("title", problem.Title)

		if problem.Type != "" {
			logger = logger.With("type", problem.Type)
		}

		if problem.Detail != "" {
			logger = logger.With("detail", problem.Detail)
		}

		logLevel := log.Info
		switch {
		case problem.Status >= 400 && problem.Status <= 499:
			logLevel = log.Info
		case problem.Status >= 500 && problem.Status <= 599:
			logLevel = log.Info
		}

		logger.Print(logLevel, err.Error())

		// MarshalIndent is used since problems are meant to be as human
		// readable as possible. They aren't worth minifying.
		body, _ := json.MarshalIndent(problem, "", "\t")
		w.Header().Add("Content-Type", "application/problem+json")
		w.WriteHeader(problem.Status)
		w.Write(body)
	}
}
