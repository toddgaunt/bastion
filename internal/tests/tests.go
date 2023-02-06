package tests

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

func Diff(a, b string) string {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(a, b, true)

	return dmp.DiffPrettyText(diffs)
}
