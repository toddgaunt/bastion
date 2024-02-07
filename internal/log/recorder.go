package log

import "fmt"

// Entry represents a log entry in a log recorder.
type Entry struct {
	Level Level
	Message string
}

type recorder struct {
	keyValues map[any]any
	entries    *[]Entry
}

// NewRecorder creates a new log recorder.
func NewRecorder() (Logger, *[]Entry) {
	entries := new([]Entry)

	return &recorder{
		keyValues: make(map[any]any),
		entries: entries,
	}, entries
}

// With decorates the log recorder with key:value pairs. A new logger is
// created when calling With, so the previous logger is left unmodified.
func (r *recorder) With(keyValues ...any) Logger {
	if len(keyValues)%2 != 0 {
		panic("With requires an even number of args")
	}

	n := &recorder{}
	for k, v := range r.keyValues {
		n.keyValues[k] = v
	}

	for i := 0; i < len(keyValues); i += 2 {
		n.keyValues[keyValues[i]] = keyValues[i+1]
	}

	return n
}

// Print writes a log to the log recorder.
func (r *recorder) Print(n Level, args ...any) {
	*r.entries = append(*r.entries, Entry{n, fmt.Sprint(args...)})
}

// Printf writes a log to the log recorder.
func (r *recorder) Printf(n Level, format string, args ...any) {
	*r.entries = append(*r.entries, Entry{n, fmt.Sprintf(format, args...)})
}
