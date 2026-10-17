package containers

import "iter"

// EnumerableWithIndex describes indexed traversal. Callbacks run in index order.
// Any and Find stop at the first match; All stops at the first failure.
// For an empty source Any is false and All is true. Find returns (-1, zero, false)
// on a miss. Implementations must reject nil callbacks, even when empty.
type EnumerableWithIndex[T any] interface {
	Each(func(index int, value T))
	Any(func(index int, value T) bool) bool
	All(func(index int, value T) bool) bool
	Find(func(index int, value T) bool) (int, T, bool)
}

// EnumerableWithKey has the same short-circuit rules as EnumerableWithIndex.
// Traversal order depends on the container. Find returns (zero, zero, false)
// on a miss, so legitimate zero keys and values are not mistaken for absence.
type EnumerableWithKey[K, V any] interface {
	Each(func(key K, value V))
	Any(func(key K, value V) bool) bool
	All(func(key K, value V) bool) bool
	Find(func(key K, value V) bool) (K, V, bool)
}

// Each calls action for each pair. Nil sources and callbacks panic immediately.
// Whether a sequence can be traversed repeatedly is determined by its producer.
func Each[K, V any](source iter.Seq2[K, V], action func(K, V)) {
	if source == nil || action == nil {
		panic("containers: nil sequence or action")
	}
	for key, value := range source {
		action(key, value)
	}
}

// Find returns the first matching pair, or zero values and false on a miss.
func Find[K, V any](source iter.Seq2[K, V], predicate func(K, V) bool) (K, V, bool) {
	requirePredicate(source, predicate)
	for key, value := range source {
		if predicate(key, value) {
			return key, value, true
		}
	}
	var key K
	var value V
	return key, value, false
}

// Any reports whether at least one pair matches. Empty sources return false.
func Any[K, V any](source iter.Seq2[K, V], predicate func(K, V) bool) bool {
	_, _, found := Find(source, predicate)
	return found
}

// All reports whether every pair matches. Empty sources return true.
func All[K, V any](source iter.Seq2[K, V], predicate func(K, V) bool) bool {
	requirePredicate(source, predicate)
	for key, value := range source {
		if !predicate(key, value) {
			return false
		}
	}
	return true
}

// Map lazily transforms pairs to values, allowing the output type to differ.
// A consumer's early termination stops traversal of the source.
func Map[K, V, R any](source iter.Seq2[K, V], transform func(K, V) R) iter.Seq[R] {
	if source == nil || transform == nil {
		panic("containers: nil sequence or transform")
	}
	return func(yield func(R) bool) {
		for key, value := range source {
			if !yield(transform(key, value)) {
				return
			}
		}
	}
}

// Filter lazily preserves matching pairs, including their original keys.
func Filter[K, V any](source iter.Seq2[K, V], predicate func(K, V) bool) iter.Seq2[K, V] {
	requirePredicate(source, predicate)
	return func(yield func(K, V) bool) {
		for key, value := range source {
			if predicate(key, value) && !yield(key, value) {
				return
			}
		}
	}
}

func requirePredicate[K, V any](source iter.Seq2[K, V], predicate func(K, V) bool) {
	if source == nil || predicate == nil {
		panic("containers: nil sequence or predicate")
	}
}
