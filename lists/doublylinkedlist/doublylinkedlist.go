// Package doublylinkedlist implements a bidirectional node-based list.
// Endpoint insertion is O(1) per value; indexed access searches the nearer end.
package doublylinkedlist

import (
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/lists/internal/listcore"
)

type List[T any] struct {
	listcore.Core[T, listcore.Double]
}
type Iterator[T any] = listcore.Cursor[T, listcore.Double]

func New[T any]() *List[T] { return &List[T]{} }
func NewWithComparator[T any](order compare.Comparator[T]) *List[T] {
	l := New[T]()
	l.Configure(order)
	return l
}
func (l *List[T]) Iterator() *Iterator[T] { return l.Cursor() }
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
