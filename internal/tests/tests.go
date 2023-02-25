package tests

import (
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// MockClock is a clock for testing
type MockClock time.Time

// Now returns the time p is set to return.
func (p MockClock) Now() time.Time {
	return time.Time(p)
}

func Diff(a, b string) string {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(a, b, true)

	return dmp.DiffPrettyText(diffs)
}
