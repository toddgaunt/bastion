package gmath

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Abs[T Number](a T) T {
	var zero T
	if a < zero {
		return -a
	}
	return a
}

func Concat(args ...string) string {
	var res strings.Builder
	for _, s := range args {
		fmt.Fprint(&res, s)
	}

	return res.String()
}
