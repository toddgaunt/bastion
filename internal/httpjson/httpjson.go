/*
Package httpjson is a layer for writing HTTP request handlers that return JSON
data or JSON problems.
*/
package httpjson

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	contentTypeHeader  = "Content-Type"
	contentTypeJSON    = "application/json"
	contentTypeProblem = "application/problem+json"
)

// Response represents a JSON response payload.
type Response struct {
	Status int
	Header http.Header
	Object interface{}
}

// Problem represents server errors in JSON, defined by IETF RFC 7807.
type Problem struct {
	Type     url.URL `json:"type"`
	Title    string  `json:"title"`
	Status   int     `json:"status"`
	Detail   string  `json:"detail,omitempty"`
	Instance string  `json:"instance,omitempty"`
}

func writeProblem(w http.ResponseWriter, problem Problem) {
	// Fill in values that we left unfilled
	if problem.Status == 0 {
		problem.Status = http.StatusInternalServerError
	}
	if problem.Title == "" {
		problem.Title = http.StatusText(problem.Status)
	}
	// MarshalIndent is used since problems are meant to be as human
	// readable as possible, they aren't worth minifying.
	var body, _ = json.MarshalIndent(problem, "", "  ")
	w.Header().Add(contentTypeHeader, contentTypeProblem)
	w.WriteHeader(problem.Status)
	w.Write(body)
}

// HandlerFunc wraps an HTTP handler that returns a Response and a Problem as a
// standard http.HandlerFunc that returns nothing. This allows for more
// familiar error returns within JSON based HTTP handlers while allowing these
// handlers to be compatible with functions that expect a standard
// http.HandlerFunc.
func HandlerFunc(
	jsonHandler func(r *http.Request) (*Response, *Problem),
) func(w http.ResponseWriter, r *http.Request) {
	var wrapped = func(w http.ResponseWriter, r *http.Request) {
		var response, problem = jsonHandler(r)
		if response == nil && problem == nil {
			writeProblem(w, Problem{})
		}
		if problem != nil {
			writeProblem(w, *problem)
		}

		var body, err = json.Marshal(response.Object)
		if err != nil {
			writeProblem(w, Problem{})
		}
		var header = w.Header()
		header.Add(contentTypeHeader, contentTypeJSON)
		for key, values := range response.Header {
			for _, val := range values {
				header.Add(key, val)
			}
		}
		w.WriteHeader(response.Status)
		w.Write(body)
	}

	return wrapped
}
