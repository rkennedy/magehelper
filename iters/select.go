package iters

import "iter"

// SliceSelectFirst returns the first item from the sequence for which the predicate returns true. If no item is found,
// then it returns a default value and false.
func SliceSelectFirst[T any](items iter.Seq[T], predicate func(T) bool) (T, bool) {
	for t := range items {
		if predicate(t) {
			return t, true
		}
	}
	return *new(T), false
}
