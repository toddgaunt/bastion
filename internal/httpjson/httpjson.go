/*
Package httpjson is a layer for writing HTTP request handlers that return JSON
data or JSON problems.
*/
package httpjson

import (
	"encoding/json"
	"net/http"
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
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Err      error  `json:"-"`
}

func pad(s, pad string) string {
	if s == "" {
		return s
	}

	return s + pad
}

func (p *Problem) Error() string {
	str := ""

	if p.Type != "" {
		str = pad(str, ": ")
		str += p.Type
	}

	if p.Title != "" {
		str = pad(str, ": ")
		str += p.Title
	}

	if p.Detail != "" {
		str = pad(str, ": ")
		str += p.Detail
	}

	if p.Err != nil {
		str = pad(str, ": ")
		str += p.Err.Error()
	}

	return str
}

// WriteProblem writes a Problem as an http response.
func WriteProblem(w http.ResponseWriter, problem Problem) {
	// Fill in values that we left unfilled
	if problem.Status == 0 {
		problem.Status = http.StatusInternalServerError
	}

	if problem.Title == "" {
		problem.Title = http.StatusText(problem.Status)
	}

	// MarshalIndent is used since problems are meant to be as human
	// readable as possible. They aren't worth minifying.
	body, _ := json.MarshalIndent(problem, "", "  ")
	w.Header().Add(contentTypeHeader, contentTypeProblem)
	w.WriteHeader(problem.Status)
	w.Write(body)
}

// HandlerFunc wraps an HTTP handler that returns a Response and a Problem as a
// standard http.HandlerFunc that returns nothing. This allows for values to be
// returned rather than writing the response procedurally while still being
// compatible with functions that expect a standard http.HandlerFunc. The
// response's content type is set to "application/json" if no Content-Type
// header was set in the Response.
func HandlerFunc(
	jsonHandler func(r *http.Request) (*Response, *Problem),
) func(w http.ResponseWriter, r *http.Request) {
	var wrapped = func(w http.ResponseWriter, r *http.Request) {
		response, problem := jsonHandler(r)

		if response == nil && problem == nil {
			WriteProblem(w, Problem{})
		}

		if problem != nil {
			WriteProblem(w, *problem)
		}

		body, err := json.Marshal(response.Object)
		if err != nil {
			WriteProblem(w, Problem{})
		}

		header := w.Header()
		for key, values := range response.Header {
			for _, val := range values {
				header.Add(key, val)
			}
		}

		// Set Content-Type to JSON only if there wasn't one provided by the caller.
		if header.Get(contentTypeHeader) == "" {
			header.Add(contentTypeHeader, contentTypeJSON)
		}

		w.WriteHeader(response.Status)
		w.Write(body)
	}

	return wrapped
}
