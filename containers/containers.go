// Package containers defines shared contracts and helpers for data structures.
// Capabilities are composed rather than requiring every container to implement
// mutation, iteration, equality, hashing or serialization in the same way.
// Implementations must document ordering and concurrency guarantees.
package containers

import (
	"cmp"
	"slices"

	"github.com/Shenghui-Waa/go.repo/compare"
)

// ValuesProvider exposes values in the container's iteration order, if defined.
// Helpers accept this small interface so read-only sources can also use them.
type ValuesProvider[T any] interface {
	Values() []T
}

// Container is a read-only observation contract. Empty agrees with Size.
// Values returns a shallow snapshot: callers may
// edit its slice without changing membership, but referenced objects are shared.
type Container[T any] interface {
	ValuesProvider[T]
	Empty() bool
	Size() int
}

// GetSortedValues returns an independently owned, sorted shallow snapshot.
// It never sorts the source's slice, even if Values exposes internal storage.
// Nil and empty slices retain their distinction. A nil source is invalid.
func GetSortedValues[T cmp.Ordered](source ValuesProvider[T]) []T {
	return GetSortedValuesFunc(source, compare.Ordered[T]())
}

// GetSortedValuesFunc stably sorts a shallow snapshot using order. Equal-ranked
// values retain their source order. A nil comparator panics even for empty input.
func GetSortedValuesFunc[T any](source ValuesProvider[T], order compare.Comparator[T]) []T {
	if order == nil {
		panic("containers: nil comparator")
	}
	values := slices.Clone(source.Values())
	slices.SortStableFunc(values, order)
	return values
}

// GetSortedComparableValues sorts by each element's CompareTo implementation.
func GetSortedComparableValues[T compare.Comparable[T]](source ValuesProvider[T]) []T {
	return GetSortedValuesFunc(source, compare.FromComparable[T]())
}
