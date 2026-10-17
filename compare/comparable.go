package compare

// Comparable defines a type's natural order. CompareTo returns a negative,
// zero or positive result for less than, equivalent to or greater than other.
// Implementations must define a consistent, transitive ordering.
type Comparable[T any] interface {
	CompareTo(other T) int
}

// FromComparable adapts natural ordering without runtime type assertions.
// Nil handling, if applicable, belongs to the type's CompareTo implementation.
func FromComparable[T Comparable[T]]() Comparator[T] {
	return func(left, right T) int {
		return left.CompareTo(right)
	}
}
