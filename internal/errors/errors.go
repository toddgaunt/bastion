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

// Msg is a message that can be safely passed along to clients.
type Msg string

// Key is string that is meant to be a stable way to refer to a particular
// error that clients can act upon.
type Key string

// Msgf is like fmt.Sprintf, except it returns the string as type Msg.
func Msgf(format string, args ...any) Msg {
	return Msg(fmt.Sprintf(format, args...))
}

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

// E is a custom error type which contains an operation, http status code, an
// error key, a client facing message, and an underlying error which caused it.
type E struct {
	Op   Op
	Code int
	Key  Key
	Msg  Msg
	Err  error
}

// pad adds s to msg if msg isn't empty.
func pad(msg, s string) string {
	if len(msg) == 0 {
		return msg
	}

	return msg + s
}

// Error returns a string with as much detail about the error as possible. This
// value should not be exposed to external clients.
func (e E) Error() string {
	msg := ""

	if e.Op != "" {
		msg = pad(msg, ": ")
		msg += string(e.Op)
	}

	if e.Key != "" {
		if next, ok := e.Err.(*E); ok {
			if e.Key != next.Key {
				msg = pad(msg, ": ")
				msg += string(e.Key)
			}
		} else {
			msg = pad(msg, ": ")
			msg += string(e.Key)
		}
	}

	if e.Err != nil {
		msg = pad(msg, ": ")

		if err, ok := e.Err.(*E); ok {
			msg += err.Error()
		} else {
			msg += e.Err.Error()
		}
	}

	return msg
}

// Message constructs a string from an Error's message list which is
// appropriate to return to external clients.
func (e E) Message() string {
	msg := string(e.Msg)

	if e.Err != nil {
		if err, ok := e.Err.(*E); ok {
			next := err.Message()
			if next != "" {
				msg = pad(msg, ": ")
				msg += next
			}
		}
	}

	return msg
}

// Unwrap returns the current error's underlying error, if there is one.
func (e E) Unwrap() error {
	return e.Err
}

// Is returns true if e is equivalent to the target error.
func (e E) Is(target error) bool {
	if err, ok := target.(*E); ok {
		return e.Op == err.Op && e.Code == err.Code && e.Key == err.Key
	}

	return false
}
