package httpjson

import "net/http"

// OK returns a Response with a status code of 200
func OK(object interface{}) *Response {
	return &Response{
		Status: http.StatusOK,
		Object: object,
	}
}

// Created returns a Response with a status code of 201
func Created(object interface{}) *Response {
	return &Response{
		Status: http.StatusCreated,
		Object: object,
	}
}

// Accepted returns a Response with a status code of 202
func Accepted(object interface{}) *Response {
	return &Response{
		Status: http.StatusAccepted,
		Object: object,
	}
}

// NonAuthoritativeInfo returns a Response with a status code of 203
func NonAuthoritativeInfo(object interface{}) *Response {
	return &Response{
		Status: http.StatusNonAuthoritativeInfo,
		Object: object,
	}
}

// NoContent returns a Response with a status code of 204
func NoContent(object interface{}) *Response {
	return &Response{
		Status: http.StatusNoContent,
		Object: object,
	}
}

// ResetContent returns a Response with a status code of 205
func ResetContent(object interface{}) *Response {
	return &Response{
		Status: http.StatusResetContent,
		Object: object,
	}
}

// PartialContent returns a Response with a status code of 206
func PartialContent(object interface{}) *Response {
	return &Response{
		Status: http.StatusPartialContent,
		Object: object,
	}
}

// MultiStatus returns a Response with a status code of 207
func MultiStatus(object interface{}) *Response {
	return &Response{
		Status: http.StatusMultiStatus,
		Object: object,
	}
}

// AlreadyReported returns a Response with a status code of 208
func AlreadyReported(object interface{}) *Response {
	return &Response{
		Status: http.StatusAlreadyReported,
		Object: object,
	}
}

// StatusIMUsed returns a Response with a status code of 209
func StatusIMUsed(object interface{}) *Response {
	return &Response{
		Status: http.StatusIMUsed,
		Object: object,
	}
}
