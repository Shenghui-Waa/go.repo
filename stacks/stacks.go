// Package stacks defines last-in, first-out containers.
package stacks

import (
	"fmt"

	"github.com/Shenghui-Waa/go.repo/containers"
)

// Stack permits duplicates and arbitrary values. Values and traversal are
// top-first. Implementations are not safe for concurrent access.
type Stack[T any] interface {
	containers.Container[T]
	Push(T)
	Pop() (T, bool)
	Peek() (T, bool)
}

type ExtendedStack[T any] interface {
	Stack[T]
	containers.Clearer
	containers.EnumerableWithIndex[T]
	containers.IndexedSequence[T]
	containers.JSONCodec
	fmt.Stringer
}
