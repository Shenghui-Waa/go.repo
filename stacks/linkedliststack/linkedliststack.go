// Package linkedliststack implements a node-backed LIFO stack.
package linkedliststack

import "github.com/Shenghui-Waa/go.repo/stacks/internal/stackcore"

type Stack[T any] struct {
	stackcore.Core[T, stackcore.Linked]
}
type Iterator[T any] = stackcore.Iterator[T]

func New[T any]() *Stack[T]                              { return &Stack[T]{} }
func (s *Stack[T]) Map(f func(int, T) T) *Stack[T]       { return &Stack[T]{s.Core.Map(f)} }
func (s *Stack[T]) Select(f func(int, T) bool) *Stack[T] { return &Stack[T]{s.Core.Select(f)} }
