// Package priorityqueue implements a comparator-driven binary heap.
package priorityqueue

import (
	"cmp"
	"container/heap"
	"iter"

	"github.com/Shenghui-Waa/go.repo/collections"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/iterable"
)

type PriorityQueue[T any] struct {
	values  []T
	less    compare.Comparator[T]
	equal   compare.Equal[T]
	version uint64
}

// New creates an empty priority queue using T's Java-style natural ordering.
func New[T compare.Comparable[T]]() *PriorityQueue[T] {
	return NewWithComparator(compare.FromComparable[T]())
}

// NewWithComparator creates an empty priority queue using an explicit ordering.
func NewWithComparator[T any](
	less compare.Comparator[T],
	equal ...compare.Equal[T],
) *PriorityQueue[T] {
	if less == nil {
		panic("priorityqueue: nil comparator")
	}
	equality := compare.Equal[T](func(left, right T) bool {
		return less(left, right) == 0
	})
	if len(equal) > 0 && equal[0] != nil {
		equality = equal[0]
	}
	queue := &PriorityQueue[T]{less: less, equal: equality}
	heap.Init(&heapAdapter[T]{queue: queue})
	return queue
}

func NewOrdered[T cmp.Ordered](values ...T) *PriorityQueue[T] {
	queue := NewWithComparator(compare.Natural[T])
	for _, value := range values {
		queue.Offer(value)
	}
	return queue
}

func (queue *PriorityQueue[T]) Len() int      { return len(queue.values) }
func (queue *PriorityQueue[T]) Size() int     { return queue.Len() }
func (queue *PriorityQueue[T]) IsEmpty() bool { return queue.Len() == 0 }

func (queue *PriorityQueue[T]) Offer(value T) bool {
	heap.Push(&heapAdapter[T]{queue: queue}, value)
	queue.version++
	return true
}

func (queue *PriorityQueue[T]) Peek() (T, bool) {
	if len(queue.values) == 0 {
		var zero T
		return zero, false
	}
	return queue.values[0], true
}

func (queue *PriorityQueue[T]) Poll() (T, bool) {
	if len(queue.values) == 0 {
		var zero T
		return zero, false
	}
	value := pop(queue)
	queue.version++
	return value, true
}

func pop[T any](queue *PriorityQueue[T]) T {
	return heap.Pop(&heapAdapter[T]{queue: queue}).(T)
}

func (queue *PriorityQueue[T]) Add(value T) bool { return queue.Offer(value) }

func (queue *PriorityQueue[T]) Contains(value T) bool {
	for _, current := range queue.values {
		if queue.equal(current, value) {
			return true
		}
	}
	return false
}

func (queue *PriorityQueue[T]) Remove(value T) bool {
	for index, current := range queue.values {
		if queue.equal(current, value) {
			heap.Remove(&heapAdapter[T]{queue: queue}, index)
			queue.version++
			return true
		}
	}
	return false
}

func (queue *PriorityQueue[T]) Clear() {
	if len(queue.values) == 0 {
		return
	}
	clear(queue.values)
	queue.values = nil
	queue.version++
}

// All exposes heap storage order; it is not sorted order.
func (queue *PriorityQueue[T]) All() iter.Seq[T] {
	expectedVersion := queue.version
	return func(yield func(T) bool) {
		for _, value := range queue.values {
			if expectedVersion != queue.version {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(value) {
				return
			}
		}
	}
}

type heapAdapter[T any] struct{ queue *PriorityQueue[T] }

func (adapter *heapAdapter[T]) Len() int { return len(adapter.queue.values) }
func (adapter *heapAdapter[T]) Less(i, j int) bool {
	return adapter.queue.less(adapter.queue.values[i], adapter.queue.values[j]) < 0
}
func (adapter *heapAdapter[T]) Swap(i, j int) {
	adapter.queue.values[i], adapter.queue.values[j] = adapter.queue.values[j], adapter.queue.values[i]
}
func (adapter *heapAdapter[T]) Push(value any) {
	adapter.queue.values = append(adapter.queue.values, value.(T))
}
func (adapter *heapAdapter[T]) Pop() any {
	values := adapter.queue.values
	last := len(values) - 1
	value := values[last]
	var zero T
	values[last] = zero
	adapter.queue.values = values[:last]
	return value
}

var _ collections.Queue[int] = (*PriorityQueue[int])(nil)
