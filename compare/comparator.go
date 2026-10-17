package compare

import "cmp"

// Comparator returns a negative, zero or positive result for less than,
// equivalent to or greater than. It must define a strict weak ordering, as
// required by slices.SortFunc. Ordering equivalence need not mean Go equality.
// The zero value is not usable; constructors and combinators reject nil funcs.
// As a named function type, Comparator can be passed directly to slices.SortFunc.
type Comparator[T any] func(left, right T) int

// Compare is the method form of calling the comparator.
func (c Comparator[T]) Compare(left, right T) int {
	return c(left, right)
}

// Reversed returns the opposite ordering. Swapping operands avoids overflowing
// when a comparator returns the minimum signed integer.
func (c Comparator[T]) Reversed() Comparator[T] {
	requireComparator(c)
	return func(left, right T) int {
		return c(right, left)
	}
}

// ThenComparing uses next only when this comparator returns zero.
func (c Comparator[T]) ThenComparing(next Comparator[T]) Comparator[T] {
	requireComparator(c)
	requireComparator(next)
	return func(left, right T) int {
		if result := c(left, right); result != 0 {
			return result
		}
		return next(left, right)
	}
}

// Ordered compares built-in ordered types, including named types. Like
// cmp.Compare, it orders NaN before non-NaN and treats signed zeros as equal.
func Ordered[T cmp.Ordered]() Comparator[T] {
	return cmp.Compare[T]
}

// Comparing orders values by an extracted key using the supplied key comparator.
// Each key is extracted once per comparison, not once per sort.
func Comparing[T, K any](key func(T) K, order Comparator[K]) Comparator[T] {
	if key == nil {
		panic("compare: nil key extractor")
	}
	requireComparator(order)
	return func(left, right T) int {
		return order(key(left), key(right))
	}
}

// ComparingOrdered orders values by keys with a built-in ordering.
func ComparingOrdered[T any, K cmp.Ordered](key func(T) K) Comparator[T] {
	return Comparing(key, Ordered[K]())
}

// NilFirst orders nil pointers before non-nil pointers. The supplied comparator
// receives only non-nil pointers. It must itself be non-nil.
func NilFirst[T any](order Comparator[*T]) Comparator[*T] {
	requireComparator(order)
	return func(left, right *T) int {
		switch {
		case left == nil && right == nil:
			return 0
		case left == nil:
			return -1
		case right == nil:
			return 1
		default:
			return order(left, right)
		}
	}
}

// NilLast orders nil pointers after non-nil pointers, preserving the supplied
// order for non-nil values. It has the same preconditions as NilFirst.
func NilLast[T any](order Comparator[*T]) Comparator[*T] {
	return NilFirst(order.Reversed()).Reversed()
}

func requireComparator[T any](order Comparator[T]) {
	if order == nil {
		panic("compare: nil comparator")
	}
}
