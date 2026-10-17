// Package lists defines ordered, duplicate-preserving mutable sequences.
package lists

import (
	"fmt"

	"github.com/Shenghui-Waa/go.repo/base"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
)

// List separates object equality from sorting. Invalid indexed mutations return
// false without modifying the list. Insert accepts Size(); Set does not append.
// Storage and positional operations accept any T without equality/hash checks.
// Implementations are not concurrent-safe and must not be copied after use.
type List[T any] interface {
	containers.Container[T]
	Get(int) (T, bool)
	Add(...T)
	Insert(int, ...T) bool
	Set(int, T) bool
	Remove(int) bool
}

// Searchable is optional default value-based membership and removal. Built-in
// lists require scalar values or Equaler+Hasher for operands actually compared;
// unsupported operands panic. Use IndexOfFunc/ContainsFunc for explicit equality.
type Searchable[T any] interface {
	RemoveValue(T) bool
	Contains(...T) bool
	IndexOf(T) int
	LastIndexOf(T) int
}

// Sortable supports an explicit comparator or a saved default order.
type Sortable[T any] interface {
	Sort(...compare.Comparator[T])
}

// ObjectList opts into default element-wise equality and ordered hashing.
// These methods validate operands at use, not on insertion. Unsupported values
// panic; no compile-time guarantee is implied for T any. For arbitrary elements,
// prefer EqualFunc and HashFunc with caller-provided compatible policies.
type ObjectList[T any] interface {
	List[T]
	base.Equaler
	base.Hasher
}

// ExtendedList preserves the previous all-capabilities List contract.
// Prefer List or a local composition for new consumers.
type ExtendedList[T any] interface {
	List[T]
	Searchable[T]
	Sortable[T]
	containers.Clearer
	containers.EnumerableWithIndex[T]
	containers.JSONCodec
	fmt.Stringer
	ObjectList[T]
	Append(...T)
	Prepend(...T)
	Swap(int, int) bool
}
