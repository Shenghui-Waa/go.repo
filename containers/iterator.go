package containers

import (
	"iter"
	"slices"
)

// IteratorWithIndex starts before the first element. After Next or First returns
// true, Index and Value are valid; outside that state they must panic.
// Next at exhaustion stays exhausted. Begin resets to before the first element.
// NextTo searches strictly after the current element and rejects nil predicates.
// Unless documented otherwise, structural mutation invalidates the iterator;
// callers must not mutate during traversal or assume concurrent access is safe.
type IteratorWithIndex[T any] interface {
	Next() bool
	Value() T
	Index() int
	Begin()
	First() bool
	NextTo(func(index int, value T) bool) bool
}

// IteratorWithKey follows IteratorWithIndex's cursor and invalidation rules.
// Key is valid only when Value is valid. Traversal order is container-specific.
type IteratorWithKey[K, V any] interface {
	Next() bool
	Value() V
	Key() K
	Begin()
	First() bool
	NextTo(func(key K, value V) bool) bool
}

// ReverseIteratorWithIndex also traverses backwards. End positions after the
// last element, Last selects the last element, and PrevTo searches strictly
// before the current position. Failed reverse traversal stays before the first.
// Prev after forward exhaustion selects the last element, if one exists.
type ReverseIteratorWithIndex[T any] interface {
	IteratorWithIndex[T]
	Prev() bool
	End()
	Last() bool
	PrevTo(func(index int, value T) bool) bool
}

// ReverseIteratorWithKey applies the same reverse cursor rules to keyed values.
type ReverseIteratorWithKey[K, V any] interface {
	IteratorWithKey[K, V]
	Prev() bool
	End()
	Last() bool
	PrevTo(func(key K, value V) bool) bool
}

// Indexed adapts a cursor to a single-pass sequence starting at its next item.
// It does not reset the cursor. Early termination leaves it on the last yielded
// item; another traversal continues after that item. Do not nest traversals of
// the same cursor. The cursor must be non-nil.
func Indexed[T any](cursor IteratorWithIndex[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for cursor.Next() {
			if !yield(cursor.Index(), cursor.Value()) {
				return
			}
		}
	}
}

// Keyed adapts a keyed cursor with the same consumption semantics as Indexed.
func Keyed[K, V any](cursor IteratorWithKey[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for cursor.Next() {
			if !yield(cursor.Key(), cursor.Value()) {
				return
			}
		}
	}
}

// Snapshot captures a shallow copy now and returns a repeatable indexed
// sequence. Later changes to source membership do not affect the sequence.
func Snapshot[T any](source ValuesProvider[T]) iter.Seq2[int, T] {
	return slices.All(slices.Clone(source.Values()))
}
