// Package sets defines duplicate-free collections with explicit membership rules.
package sets

import (
	"fmt"

	"github.com/Shenghui-Waa/go.repo/containers"
)

// Set uses hash/equality or comparator equivalence according to its implementation.
// Constructors are empty; Add and Remove report how many members changed.
type Set[T any] interface {
	containers.Container[T]
	Add(...T) int
	Remove(...T) int
	Contains(...T) bool
}

// ExtendedSet preserves the former all-capabilities Set contract.
type ExtendedSet[T any] interface {
	Set[T]
	containers.Clearer
	containers.EnumerableWithIndex[T]
	containers.IndexedSequence[T]
	containers.JSONCodec
	fmt.Stringer
}
