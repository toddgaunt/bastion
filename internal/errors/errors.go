// Package errors is a custom error package meant as a drop-in replacement for
// the standard errors package that allows for separation between client facing
// error messages and more detailed internal errors.
package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
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

// ModulePrefix is set during compilation to the module root to make error
// locations relative rather than absolute.
var ModulePrefix string

type Error interface {
	OpHolder
	TypeHolder
	TitleHolder
	StatusHolder
	DetailHolder
	FieldsHolder
	LocationHolder
	error
}

type Problem interface {
	StatusHolder
	DetailHolder
	error
}

type FieldsHolder interface {
	Fields() map[string]any
	error
}

type DetailHolder interface {
	Detail() string
	error
}

type StatusHolder interface {
	Status() int
	error
}

type TitleHolder interface {
	Title() string
	error
}

type TypeHolder interface {
	Type() string
	error
}

type OpHolder interface {
	Op() string
	error
}

type LocationHolder interface {
	Location() Location
	error
}

type Location struct {
	file string
	line int
}

// Note is a set of fields that can be filled to wrap an error with.
type Note struct {
	Op         Op
	Type       Type
	Title      Title
	StatusCode int
	Detail     string
	Fields     map[string]any
}

func getLocation() (string, int) {
	_, file, line, _ := runtime.Caller(2)

	fmt.Println("Module prefix:", ModulePrefix)
	file, _ = strings.CutPrefix(file, ModulePrefix)

	return file, line
}

// Wrap annotates err with the values present in the note.
//
// The calling function's file and line number is recorded
// in the returned error and can be retrieved by calling
// Error.Location()
func (n Note) Wrap(err error) Error {
	file, line := getLocation()

	return annotatedError{n, file, line, err}
}

// Wraps creates an error from the provided string and annotates it
// with the values present in the note.
//
// See Wrap() for more details.
func (n Note) Wraps(msg string) Error {
	file, line := getLocation()

	return annotatedError{n, file, line, New(msg)}
}

// Wrapf creates an error from the format string and arguments and
// annotates it with the values present in the note.
func (n Note) Wrapf(format string, args ...any) Error {
	file, line := getLocation()

	return annotatedError{n, file, line, Errorf(format, args...)}
}

type annotatedError struct {
	note Note
	file string
	line int
	err  error
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

	if e.note.Op != "" {
		msg = pad(msg, ": ")
		msg += string(e.note.Op)
	}

	if e.note.Type != "" {
		msg = pad(msg, ": ")
		msg += string(e.note.Type)
	}

	if e.note.Title != "" {
		msg = pad(msg, ": ")
		msg += string(e.note.Title)
	}

	if e.note.StatusCode != 0 {
		msg = pad(msg, ": ")
		msg += fmt.Sprintf("%d", e.note.StatusCode)
	}

	if e.err != nil {
		msg = pad(msg, ": ")
		msg += e.err.Error()
	}

	return msg
}

// Type returns the kind or class of the error.
func (e annotatedError) Op() string {
	if e.note.Op == "" {
		if err, ok := e.err.(OpHolder); ok {
			return string(err.Op())
		}
	}

	return string(e.note.Op)
}

// Type returns the kind or class of the error.
func (e annotatedError) Type() string {
	if e.note.Type == "" {
		if err, ok := e.err.(TypeHolder); ok {
			return err.Type()
		}
	}

	return string(e.note.Type)
}

// Title returns the human readble title of an error.
func (e annotatedError) Title() string {
	if e.note.Title == "" {
		if err, ok := e.err.(TitleHolder); ok {
			return err.Title()
		}
	}

	return string(e.note.Title)
}

// Status returns the error's status code.
func (e annotatedError) Status() int {
	if e.note.StatusCode == 0 {
		if err, ok := e.err.(StatusHolder); ok {
			return err.Status()
		}
	}

	return e.note.StatusCode
}

// Detail constructs a string from an Error's message list which is
// appropriate to return to external clients.
func (e annotatedError) Detail() string {
	msg := string(e.note.Detail)

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

// Fields returns a map of all fields associated with the chain of errors. If a
// field in an error matches a field in a wrapped error, the field is
// transformed into a slice and will include the wrapped value.
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

	for k, v := range e.note.Fields {
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

func (e annotatedError) Location() Location {
	return Location{e.file, e.line}
}

// Unwrap returns the current error's underlying error, if there is one.
func (e annotatedError) Unwrap() error {
	return e.err
}

// Is returns true if target has the same type as e.
func (e annotatedError) Is(target error) bool {
	if target, ok := target.(interface{ Type() Type }); ok {
		return e.note.Type == target.Type()
	}

	return false
}

func LocationList(err error) []Location {
	var locationList []Location
	var walk = err
	for walk != nil {
		if l, ok := walk.(LocationHolder); ok {
			locationList = append(locationList, l.Location())
		}
		walk = errors.Unwrap(walk)
	}

	return locationList
}
