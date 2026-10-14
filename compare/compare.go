// Package compare contains reusable ordering and equality strategies.
package compare

import "cmp"

// Comparator returns a negative value when left precedes right, zero when they
// are equivalent for ordering, and a positive value otherwise.
type Comparator[T any] func(left, right T) int

// Equal compares two values for semantic equality.
type Equal[T any] func(left, right T) bool

// Natural returns the natural ordering for an ordered Go value.
func Natural[T cmp.Ordered](left, right T) int {
	return cmp.Compare(left, right)
}

// Reverse reverses an ordering strategy.
func Reverse[T any](comparator Comparator[T]) Comparator[T] {
	return func(left, right T) int {
		return comparator(right, left)
	}
}

// NativeEqual compares values that satisfy Go's comparable constraint.
func NativeEqual[T comparable](left, right T) bool {
	return left == right
}

// Comparable is the Java-style comparison contract.
type Comparable[T any] interface {
	CompareTo(other T) int
}

// FromComparable adapts a Java-style Comparable value to a Comparator.
func FromComparable[T Comparable[T]]() Comparator[T] {
	return func(left, right T) int {
		return left.CompareTo(right)
	}
}
