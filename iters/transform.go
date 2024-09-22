package iters

import "iter"

// SliceTransform applies a function f to each item in a source sequence and returns a sequence of the results.
func SliceTransform[T, U any](src iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for t := range src {
			if !yield(f(t)) {
				return
			}
		}
	}
}

// SliceTransform2 applies a function f to each item in a source sequence and returns a sequence of key-value results.
func SliceTransform2[T any, K comparable, V any](src iter.Seq[T], f func(T) (K, V)) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for t := range src {
			if !yield(f(t)) {
				return
			}
		}
	}
}

// MapTransform applies a function f to each key-value pair in a source sequence and returns a sequence of the results.
func MapTransform[K comparable, V, T any](src iter.Seq2[K, V], f func(K, V) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for k, v := range src {
			if !yield(f(k, v)) {
				return
			}
		}
	}
}

// MapTransform2 applies a function f to each key-value pair in a source sequence and returns a sequence of key-value
// results.
func MapTransform2[K comparable, V any, K2 comparable, V2 any](
	src iter.Seq2[K, V],
	f func(K, V) (K2, V2),
) iter.Seq2[K2, V2] {
	return func(yield func(K2, V2) bool) {
		for k, v := range src {
			if !yield(f(k, v)) {
				return
			}
		}
	}
}
