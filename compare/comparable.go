package compare

import "github.com/emirpasic/gods/utils"

// Comparable is the Java-style comparison contract.
type Comparable[T any] interface {
	CompareTo(other T) int
}

// FromComparable adapts the Java-style CompareTo contract to GoDS'
// interface-based comparator contract.
//
// The type assertions are intentionally kept at this compatibility boundary:
// GoDS v1.18.1 defines utils.Comparator as func(interface{}, interface{}) int.
func FromComparable[T Comparable[T]]() utils.Comparator {
	return func(left, right interface{}) int {
		return left.(T).CompareTo(right.(T))
	}
}
