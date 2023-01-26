// Package errors is a custom error package meant as a drop-in replacement for
// the standard errors package that allows for separation between client facing
// error messages and more detailed internal errors.
package errors

import (
	"errors"
	"fmt"
)

// Op describes an operation, usually as the http.method or logical operation.
type Op string

// Type indicates the kind or class of error encountered. It is also an error
// itself, suitable for creating sentinel errors.
type Type string

// Error returns the type as an error string.
func (t Type) Error() string {
	return string(t)
}

// Type is an identity function.
func (t Type) Type() Type {
	return t
}

// Title is a human readable title for an error.
type Title string

// As finds the first error in err's chain that matches target, and if one is
// found, sets target to that error value and returns true. Otherwise, it
// returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained
// by repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the
// value pointed to by target, or if the error has a method As(interface{})
// bool such that As(target) returns true. In the latter case, the As method is
// responsible for setting target.
//
// An error type might provide an As method so it can be treated as if it were
// a different error type.
//
// As panics if target is not a non-nil pointer to either a type that
// implements error, or to any interface type.
var As = errors.As

// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// An error type might provide an Is method so it can be treated as equivalent
// to an existing error. For example, if MyError defines
//
//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
//
// then Is(MyError{}, fs.ErrExist) returns true. See syscall.Errno.Is for
// an example in the standard library. An Is method should only shallowly
// compare err and the target and not call Unwrap on either.
var Is = errors.Is

// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
var New = errors.New

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
var Unwrap = errors.Unwrap

// Errorf formats according to a format specifier and returns the string as a
// value that satisfies error.
//
// If the format specifier includes a %w verb with an error operand,
// the returned error will implement an Unwrap method returning the operand. It is
// invalid to include more than one %w verb or to supply it with an operand
// that does not implement the error interface. The %w verb is otherwise
// a synonym for %v.
var Errorf = fmt.Errorf

// Annotation is a set of fields that can be filled to wrap an error with.
type Annotation struct {
	WithOp     Op
	WithType   Type
	WithTitle  Title
	WithStatus int
	WithDetail string
}

// Wrap annotates err with the values present in the annotation.
func (a Annotation) Wrap(err error) error {
	return annotatedError{a, err}
}

type annotatedError struct {
	Annotation
	err error
}

// pad adds s to msg if msg isn't empty.
func pad(msg, s string) string {
	if len(msg) == 0 {
		return msg
	}

	return msg + s
}

// Error returns the underlying error detail prefixed with the operation if provided.
func (e annotatedError) Error() string {
	msg := ""

	if e.WithOp != "" {
		msg = pad(msg, ": ")
		msg += string(e.WithOp)
	}

	if e.err != nil {
		msg = pad(msg, ": ")
		msg += e.err.Error()
	}

	return msg
}

// Type returns the kind or class of the error.
func (e annotatedError) Type() string {
	if e.WithType == "" {
		if err, ok := e.err.(interface{ Type() string }); ok {
			return err.Type()
		}
	}

	return string(e.WithType)
}

// Title returns the human readble title of an error.
func (e annotatedError) Title() string {
	if e.WithTitle == "" {
		if err, ok := e.err.(interface{ Title() string }); ok {
			return err.Title()
		}
	}

	return string(e.WithTitle)
}

// Status returns the error's status code.
func (e annotatedError) Status() int {
	if e.WithStatus == 0 {
		if err, ok := e.err.(interface{ Status() int }); ok {
			return err.Status()
		}
	}

	return e.WithStatus
}

// Detail constructs a string from an Error's message list which is
// appropriate to return to external clients.
func (e annotatedError) Detail() string {
	msg := string(e.WithDetail)

	if e.err != nil {
		if err, ok := e.err.(interface{ Detail() string }); ok {
			next := err.Detail()
			if next != "" {
				msg = pad(msg, ": ")
				msg += next
			}
		}
	}

	return msg
}

// Unwrap returns the current error's underlying error, if there is one.
func (e annotatedError) Unwrap() error {
	return e.err
}

// Is returns true if target has the same type as e.
func (e annotatedError) Is(target error) bool {
	if target, ok := target.(interface{ Type() Type }); ok {
		return e.WithType == target.Type()
	}

	return false
}
