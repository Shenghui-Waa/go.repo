package queuecore

import (
	"encoding/json"
	"fmt"
	"iter"

	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/trees/binaryheap"
)

const DefaultCapacity = 16

type Kind interface{ kind() int }
type Array struct{}
type Linked struct{}
type Circular struct{}
type Priority struct{}

func (Array) kind() int    { return 0 }
func (Linked) kind() int   { return 1 }
func (Circular) kind() int { return 2 }
func (Priority) kind() int { return 3 }

// Core provides shared queue contracts without exposing mutation of the storage.
// Do not copy after first use. Values and traversal are shallow snapshots.
type Core[T any, K Kind] struct {
	data     storage[T]
	order    compare.Comparator[T]
	capacity int
	version  uint64
}

func New[T any, K Kind](order compare.Comparator[T], capacity int) Core[T, K] {
	result := Core[T, K]{order: order, capacity: capacity}
	result.ready()
	return result
}
func (q *Core[T, K]) ready() {
	if q.data != nil {
		return
	}
	var kind K
	switch kind.kind() {
	case 0:
		q.data = &ring[T]{}
	case 1:
		q.data = &linked[T]{}
	case 2:
		if q.capacity == 0 {
			q.capacity = DefaultCapacity
		}
		if q.capacity < 1 {
			panic("circularbuffer: capacity must be positive")
		}
		q.data = &ring[T]{buffer: make([]T, q.capacity), fixed: true}
	case 3:
		q.data = &priority[T]{heap: binaryheap.NewWithComparator(q.order)}
	}
}
func (q *Core[T, K]) Enqueue(value T) {
	q.ready()
	// Invalidate before a callback can panic after partially modifying a heap.
	q.version++
	q.data.push(value)
}
func (q *Core[T, K]) Dequeue() (T, bool) {
	q.ready()
	if q.data.len() == 0 {
		var zero T
		return zero, false
	}
	q.version++
	return q.data.pop()
}
func (q *Core[T, K]) Peek() (T, bool) { q.ready(); return q.data.peek() }
func (q *Core[T, K]) Size() int       { q.ready(); return q.data.len() }
func (q *Core[T, K]) Empty() bool     { return q.Size() == 0 }
func (q *Core[T, K]) Clear() {
	if !q.Empty() {
		q.data.clear()
		q.version++
	}
}
func (q *Core[T, K]) Capacity() int    { q.ready(); return q.capacity }
func (q *Core[T, K]) Full() bool       { return q.Capacity() > 0 && q.Size() == q.capacity }
func (q *Core[T, K]) Values() []T      { q.ready(); return q.data.values() }
func (q *Core[T, K]) String() string   { return fmt.Sprint(q.Values()) }
func (q *Core[T, K]) ToString() string { return q.String() }
func (q *Core[T, K]) check(version uint64) {
	if q.version != version {
		panic("queues: traversal invalidated by mutation")
	}
}

// Seq captures a shallow snapshot and source version at creation. Repeated
// traversal is allowed until mutation; early termination propagates immediately.
func (q *Core[T, K]) Seq() iter.Seq2[int, T] {
	values := q.Values()
	version := q.version
	return func(yield func(int, T) bool) {
		q.check(version)
		for i, value := range values {
			more := yield(i, value)
			q.check(version)
			if !more {
				return
			}
		}
	}
}
func (q *Core[T, K]) Each(f func(int, T))          { containers.Each(q.Seq(), f) }
func (q *Core[T, K]) Any(f func(int, T) bool) bool { return containers.Any(q.Seq(), f) }
func (q *Core[T, K]) All(f func(int, T) bool) bool { return containers.All(q.Seq(), f) }
func (q *Core[T, K]) Find(f func(int, T) bool) (int, T, bool) {
	i, value, ok := containers.Find(q.Seq(), f)
	if !ok {
		i = -1
	}
	return i, value, ok
}
func (q *Core[T, K]) fresh() Core[T, K] { q.ready(); return New[T, K](q.order, q.capacity) }
func (q *Core[T, K]) Map(f func(int, T) T) Core[T, K] {
	if f == nil {
		panic("queues: nil transform")
	}
	result := q.fresh()
	q.Each(func(i int, value T) { result.Enqueue(f(i, value)) })
	return result
}
func (q *Core[T, K]) Select(f func(int, T) bool) Core[T, K] {
	if f == nil {
		panic("queues: nil predicate")
	}
	result := q.fresh()
	q.Each(func(i int, value T) {
		if f(i, value) {
			result.Enqueue(value)
		}
	})
	return result
}
func (q *Core[T, K]) MarshalJSON() ([]byte, error) { return json.Marshal(q.Values()) }
func (q *Core[T, K]) ToJSON() ([]byte, error)      { return q.MarshalJSON() }
func (q *Core[T, K]) FromJSON(data []byte) error   { return q.UnmarshalJSON(data) }
func (q *Core[T, K]) UnmarshalJSON(data []byte) error {
	var values []T
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	result := q.fresh()
	for _, value := range values {
		result.Enqueue(value)
	}
	result.version = q.version + 1
	*q = result
	return nil
}
