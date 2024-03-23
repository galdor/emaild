package utils

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Integer | constraints.Float
}

// I wish it was a joke, but no, Go does not have an absolute value function for
// integers. Do not get me started on that...
func Abs[T Number](n T) T {
	var zero T
	if n < zero {
		return -n
	}

	return n
}
