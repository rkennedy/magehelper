package iters

import "iter"

// Filter returns a new sequence based on seq consisting only of those items where predicate returns true.
func Filter[T any](seq iter.Seq[T], predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range seq {
			if predicate(item) && !yield(item) {
				return
			}
		}
	}
}

// Filter2 returns a new sequence based on seq consisting only of those pairs of items where predicate returns true.
func Filter2[K comparable, V any](seq iter.Seq2[K, V], predicate func(K, V) bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range seq {
			if predicate(k, v) && !yield(k, v) {
				return
			}
		}
	}
}
