// Package circularbuffer provides bounded FIFO storage that overwrites oldest
// values when full, matching GoDS's buffer semantics.
package circularbuffer

import "github.com/Shenghui-Waa/go.repo/queues/internal/queuecore"

const DefaultCapacity = queuecore.DefaultCapacity

type Queue[T any] struct {
	queuecore.Core[T, queuecore.Circular]
}
type Iterator[T any] = queuecore.Iterator[T]

func New[T any]() *Queue[T] { return NewWithCapacity[T](DefaultCapacity) }
func NewWithCapacity[T any](capacity int) *Queue[T] {
	if capacity < 1 {
		panic("circularbuffer: capacity must be positive")
	}
	return &Queue[T]{queuecore.New[T, queuecore.Circular](nil, capacity)}
}
func (q *Queue[T]) Map(f func(int, T) T) *Queue[T]       { return &Queue[T]{q.Core.Map(f)} }
func (q *Queue[T]) Select(f func(int, T) bool) *Queue[T] { return &Queue[T]{q.Core.Select(f)} }
