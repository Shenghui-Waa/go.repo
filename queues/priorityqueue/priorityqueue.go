// Package priorityqueue dequeues the minimum element in comparator order.
package priorityqueue

import (
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/queues/internal/queuecore"
)

type Queue[T any] struct {
	queuecore.Core[T, queuecore.Priority]
}
type Iterator[T any] = queuecore.Iterator[T]

func New[T compare.Comparable[T]]() *Queue[T] {
	return NewWithComparator(compare.FromComparable[T]())
}
func NewWithComparator[T any](order compare.Comparator[T]) *Queue[T] {
	return &Queue[T]{queuecore.New[T, queuecore.Priority](order, 0)}
}
func (q *Queue[T]) Map(f func(int, T) T) *Queue[T]       { return &Queue[T]{q.Core.Map(f)} }
func (q *Queue[T]) Select(f func(int, T) bool) *Queue[T] { return &Queue[T]{q.Core.Select(f)} }
