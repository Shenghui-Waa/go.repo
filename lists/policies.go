package lists

import "github.com/Shenghui-Waa/go.repo/containers"

// EqualFunc compares shallow snapshots in order using only equal. No Base,
// Equaler or Hasher contract is required. A nil callback panics even if empty.
// Callbacks must not mutate the sources. Nil sources are invalid.
func EqualFunc[T any](left, right containers.ValuesProvider[T], equal func(T, T) bool) bool {
	if equal == nil {
		panic("lists: nil equality function")
	}
	a, b := left.Values(), right.Values()
	if len(a) != len(b) {
		return false
	}
	for i, value := range a {
		if !equal(value, b[i]) {
			return false
		}
	}
	return true
}

// HashFunc combines caller-provided element hashes in order. The empty hash is
// 1. Equal elements must hash equally when used with EqualFunc. No reflection or
// default object hashing is used. The callback must not mutate the source.
func HashFunc[T any](source containers.ValuesProvider[T], hash func(T) uint64) uint64 {
	if hash == nil {
		panic("lists: nil hash function")
	}
	result := uint64(1)
	for _, value := range source.Values() {
		result = 31*result + hash(value)
	}
	return result
}

// IndexOfFunc returns the first equal element's index or -1. Equality is explicit
// and independent of sorting or hashing. The callback must not mutate source.
func IndexOfFunc[T any](source containers.ValuesProvider[T], value T, equal func(T, T) bool) int {
	if equal == nil {
		panic("lists: nil equality function")
	}
	for i, current := range source.Values() {
		if equal(current, value) {
			return i
		}
	}
	return -1
}

// ContainsFunc reports whether value matches using the explicit equality policy.
func ContainsFunc[T any](source containers.ValuesProvider[T], value T, equal func(T, T) bool) bool {
	return IndexOfFunc(source, value, equal) >= 0
}
