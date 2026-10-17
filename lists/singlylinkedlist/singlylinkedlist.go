// Package singlylinkedlist implements a forward-linked list with a tail pointer.
// Append is O(1) per value; indexed access and predecessor lookup are O(n).
package singlylinkedlist

import (
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/lists/internal/listcore"
)

type List[T any] struct {
	listcore.Core[T, listcore.Single]
}

// Iterator deliberately exposes forward traversal only.
type Iterator[T any] struct {
	containers.IteratorWithIndex[T]
}

func New[T any]() *List[T] { return &List[T]{} }
func NewWithComparator[T any](order compare.Comparator[T]) *List[T] {
	l := New[T]()
	l.Configure(order)
	return l
}
func (l *List[T]) Iterator() *Iterator[T] { return &Iterator[T]{l.Cursor()} }
func (l *List[T]) Map(f func(int, T) T) *List[T] {
	if f == nil {
		panic("lists: nil transform")
	}
	result := New[T]()
	if l.Comparator() != nil {
		result.Configure(l.Comparator())
	}
	l.Each(func(i int, value T) { result.Add(f(i, value)) })
	return result
}
func (l *List[T]) Select(f func(int, T) bool) *List[T] {
	if f == nil {
		panic("lists: nil predicate")
	}
	result := New[T]()
	if l.Comparator() != nil {
		result.Configure(l.Comparator())
	}
	l.Each(func(i int, value T) {
		if f(i, value) {
			result.Add(value)
		}
	})
	return result
}
