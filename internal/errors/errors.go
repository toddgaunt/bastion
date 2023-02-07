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

// itle is a human readable title for an error.
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

type Error interface {
	OpHolder
	TypeHolder
	TitleHolder
	StatusHolder
	DetailHolder
	FieldsHolder
	error
}

type Problem interface {
	StatusHolder
	DetailHolder
	error
}

type FieldsHolder interface {
	Fields() map[string]any
}

type DetailHolder interface {
	Detail() string
}

type StatusHolder interface {
	Status() int
}

type TitleHolder interface {
	Title() string
}

type TypeHolder interface {
	Type() string
}

type OpHolder interface {
	Op() string
}

// Annotation is a set of fields that can be filled to wrap an error with.
type Annotation struct {
	WithOp     Op
	WithType   Type
	WithTitle  Title
	WithStatus int
	WithDetail string
	WithFields map[string]any
}

// Wrap annotates err with the values present in the annotation.
func (a Annotation) Wrap(err error) Error {
	if err == nil {
		return nil
	}
	return annotatedError{a, err}
}

// Wrapf annotates err with the values present in the annotation.
func (a Annotation) Wrapf(format string, args ...any) Error {
	return annotatedError{a, Errorf(format, args...)}
}

type annotatedError struct {
	ann Annotation
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

	if e.ann.WithOp != "" {
		msg = pad(msg, ": ")
		msg += string(e.ann.WithOp)
	}

	if e.ann.WithType != "" {
		msg = pad(msg, ": ")
		msg += string(e.ann.WithType)
	}

	if e.ann.WithTitle != "" {
		msg = pad(msg, ": ")
		msg += string(e.ann.WithTitle)
	}

	if e.ann.WithStatus != 0 {
		msg = pad(msg, ": ")
		msg += fmt.Sprintf("%d", e.ann.WithStatus)
	}

	if e.err != nil {
		msg = pad(msg, ": ")
		msg += e.err.Error()
	}

	return msg
}

// Type returns the kind or class of the error.
func (e annotatedError) Op() string {
	if e.ann.WithOp == "" {
		if err, ok := e.err.(OpHolder); ok {
			return string(err.Op())
		}
	}

	return string(e.ann.WithOp)
}

// Type returns the kind or class of the error.
func (e annotatedError) Type() string {
	if e.ann.WithType == "" {
		if err, ok := e.err.(TypeHolder); ok {
			return err.Type()
		}
	}

	return string(e.ann.WithType)
}

// Title returns the human readble title of an error.
func (e annotatedError) Title() string {
	if e.ann.WithTitle == "" {
		if err, ok := e.err.(TitleHolder); ok {
			return err.Title()
		}
	}

	return string(e.ann.WithTitle)
}

// Status returns the error's status code.
func (e annotatedError) Status() int {
	if e.ann.WithStatus == 0 {
		if err, ok := e.err.(StatusHolder); ok {
			return err.Status()
		}
	}

	return e.ann.WithStatus
}

// Detail constructs a string from an Error's message list which is
// appropriate to return to external clients.
func (e annotatedError) Detail() string {
	msg := string(e.ann.WithDetail)

	if e.err != nil {
		if err, ok := e.err.(DetailHolder); ok {
			next := err.Detail()
			if next != "" {
				msg = pad(msg, ": ")
				msg += next
			}
		}
	}

	return msg
}

// Fields returns a map of all fields associated with the chain of errors. If a field in an error
// matches a field in a wrapped error, the field is transformed into a slice and will include
// the wrapped value.
func (e annotatedError) Fields() map[string]any {
	fields := make(map[string]any)

	if e.err != nil {
		if err, ok := e.err.(FieldsHolder); ok {
			merge := err.Fields()
			for k, v := range merge {
				if s, ok := fields[k]; ok {
					fields[k] = []any{v, s}
				} else {
					fields[k] = v
				}
			}
		}
	}

	for k, v := range e.ann.WithFields {
		if s, ok := fields[k]; ok {
			fields[k] = []any{v, s}
		} else {
			fields[k] = v
		}
	}

	if len(fields) > 0 {
		return fields
	}

	return nil
}

// Unwrap returns the current error's underlying error, if there is one.
func (e annotatedError) Unwrap() error {
	return e.err
}

// Is returns true if target has the same type as e.
func (e annotatedError) Is(target error) bool {
	if target, ok := target.(interface{ Type() Type }); ok {
		return e.ann.WithType == target.Type()
	}

	return false
}
