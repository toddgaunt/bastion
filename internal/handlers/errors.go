package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/toddgaunt/bastion/internal/errors"
	"github.com/toddgaunt/bastion/internal/log"
)

var statusInternal = errors.Note{StatusCode: http.StatusInternalServerError}
var statusBadRequest = errors.Note{StatusCode: http.StatusBadRequest}
var statusUnauthorized = errors.Note{StatusCode: http.StatusUnauthorized}

// handleError takes an error from a function and extracts out any annotations
// from it to decorate a logger and fill out a ProblemJSON response.
func handleError(w http.ResponseWriter, err errors.Problem, logger log.Logger) {
	if err == nil {
		return
	}

	if err, ok := err.(errors.FieldsHolder); ok {
		for k, v := range err.Fields() {
			logger = logger.With(k, v)
		}
	}

	// problemJSON represents server errors in JSON, defined by IETF RFC 7807.
	type problemJSON struct {
		Type     string `json:"type,omitempty"`
		Title    string `json:"title,omitempty"`
		Status   int    `json:"status"`
		Detail   string `json:"detail,omitempty"`
		Instance string `json:"instance,omitempty"`
	}

	op := ""
	problem := problemJSON{}

	if err, ok := err.(errors.OpHolder); ok {
		op = err.Op()
	}

	if err, ok := err.(errors.TypeHolder); ok {
		problem.Type = err.Type()
	}

	if err, ok := err.(errors.TitleHolder); ok {
		problem.Title = err.Title()
	}

	if err, ok := err.(errors.StatusHolder); ok {
		problem.Status = err.Status()
	}

	if err, ok := err.(errors.DetailHolder); ok {
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

	if op != "" {
		logger = logger.With("op", op)
	}

	if problem.Type != "" {
		logger = logger.With("type", problem.Type)
	}

	logger = logger.With("title", problem.Title)

	logger = logger.With("status", problem.Status)

	if problem.Detail != "" {
		logger = logger.With("detail", problem.Detail)
	}

	logger = logger.With("trace", fmt.Sprintf("%v", errors.LocationList(err)))

	logLevel := log.Info
	switch {
	case problem.Status >= 400 && problem.Status <= 499:
		logLevel = log.Info
	case problem.Status >= 500 && problem.Status <= 599:
		logLevel = log.Error
	}

	logger.Print(logLevel, err.Error())

	body, _ := json.Marshal(problem)

	w.Header().Add("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	w.Write(body)
}
