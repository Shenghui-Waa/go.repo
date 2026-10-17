// Package arrayqueue implements an amortized O(1) growing ring FIFO.
package arrayqueue

import "github.com/Shenghui-Waa/go.repo/queues/internal/queuecore"

type Queue[T any] struct {
	queuecore.Core[T, queuecore.Array]
}
type Iterator[T any] = queuecore.Iterator[T]

func New[T any]() *Queue[T]                              { return &Queue[T]{} }
func (q *Queue[T]) Map(f func(int, T) T) *Queue[T]       { return &Queue[T]{q.Core.Map(f)} }
func (q *Queue[T]) Select(f func(int, T) bool) *Queue[T] { return &Queue[T]{q.Core.Select(f)} }
