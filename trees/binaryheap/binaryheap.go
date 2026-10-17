// Package binaryheap provides an array-backed min heap in comparator order.
package binaryheap

import (
	"encoding/json"
	"fmt"
	"iter"
	"slices"

	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
)

type Heap[T any] struct {
	values  []T
	order   compare.Comparator[T]
	version uint64
}

func New[T compare.Comparable[T]]() *Heap[T] {
	return NewWithComparator(compare.FromComparable[T]())
}
func NewWithComparator[T any](order compare.Comparator[T]) *Heap[T] {
	if order == nil {
		panic("binaryheap: nil comparator")
	}
	return &Heap[T]{order: order}
}
func (h *Heap[T]) ready() {
	if h.order == nil {
		panic("binaryheap: use a constructor")
	}
}
func (h *Heap[T]) Size() int   { return len(h.values) }
func (h *Heap[T]) Empty() bool { return h.Size() == 0 }
func (h *Heap[T]) Push(values ...T) {
	h.ready()
	for _, value := range values {
		if h.order(value, value) != 0 {
			panic("binaryheap: non-reflexive comparator")
		}
	}
	if len(values) == 0 {
		return
	}
	h.version++
	h.values = append(h.values, values...)
	if len(values) == 1 {
		for child := h.Size() - 1; child > 0; {
			parent := (child - 1) / 2
			if h.order(h.values[parent], h.values[child]) <= 0 {
				break
			}
			h.values[parent], h.values[child] = h.values[child], h.values[parent]
			child = parent
		}
	} else {
		h.heapify()
	}
}
func (h *Heap[T]) down(parent int) {
	for parent < len(h.values)/2 {
		child := 2*parent + 1
		if child+1 < h.Size() && h.order(h.values[child+1], h.values[child]) < 0 {
			child++
		}
		if h.order(h.values[parent], h.values[child]) <= 0 {
			return
		}
		h.values[parent], h.values[child] = h.values[child], h.values[parent]
		parent = child
	}
}
func (h *Heap[T]) heapify() {
	for i := h.Size()/2 - 1; i >= 0; i-- {
		h.down(i)
	}
}
func (h *Heap[T]) Peek() (T, bool) {
	h.ready()
	if h.Empty() {
		var zero T
		return zero, false
	}
	return h.values[0], true
}
func (h *Heap[T]) Pop() (T, bool) {
	value, ok := h.Peek()
	if !ok {
		return value, false
	}
	h.version++
	last := h.Size() - 1
	h.values[0] = h.values[last]
	var zero T
	h.values[last] = zero
	h.values = h.values[:last]
	h.down(0)
	return value, true
}
func (h *Heap[T]) Clear() {
	h.ready()
	if !h.Empty() {
		h.values = nil
		h.version++
	}
}

// Values is heap layout, not globally sorted order.
func (h *Heap[T]) Values() []T      { h.ready(); return append([]T{}, h.values...) }
func (h *Heap[T]) String() string   { return fmt.Sprint(h.Values()) }
func (h *Heap[T]) ToString() string { return h.String() }
func (h *Heap[T]) check(version uint64) {
	if version != h.version {
		panic("binaryheap: traversal invalidated")
	}
}
func (h *Heap[T]) Seq() iter.Seq2[int, T] {
	h.ready()
	version := h.version
	return func(yield func(int, T) bool) {
		h.check(version)
		for i, value := range h.values {
			more := yield(i, value)
			h.check(version)
			if !more {
				return
			}
		}
	}
}
func (h *Heap[T]) Each(f func(int, T))          { containers.Each(h.Seq(), f) }
func (h *Heap[T]) Any(f func(int, T) bool) bool { return containers.Any(h.Seq(), f) }
func (h *Heap[T]) All(f func(int, T) bool) bool { return containers.All(h.Seq(), f) }
func (h *Heap[T]) Find(f func(int, T) bool) (int, T, bool) {
	i, value, ok := containers.Find(h.Seq(), f)
	if !ok {
		i = -1
	}
	return i, value, ok
}
func (h *Heap[T]) Map(f func(int, T) T) *Heap[T] {
	if f == nil {
		panic("binaryheap: nil transform")
	}
	result := NewWithComparator(h.order)
	values := make([]T, 0, h.Size())
	h.Each(func(i int, value T) { values = append(values, f(i, value)) })
	result.Push(values...)
	return result
}
func (h *Heap[T]) Select(f func(int, T) bool) *Heap[T] {
	if f == nil {
		panic("binaryheap: nil predicate")
	}
	result := NewWithComparator(h.order)
	values := make([]T, 0, h.Size())
	h.Each(func(i int, value T) {
		if f(i, value) {
			values = append(values, value)
		}
	})
	result.Push(values...)
	return result
}
func (h *Heap[T]) MarshalJSON() ([]byte, error) { return json.Marshal(h.Values()) }
func (h *Heap[T]) ToJSON() ([]byte, error)      { return h.MarshalJSON() }
func (h *Heap[T]) FromJSON(data []byte) error   { return h.UnmarshalJSON(data) }
func (h *Heap[T]) UnmarshalJSON(data []byte) error {
	h.ready()
	var values []T
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	result := NewWithComparator(h.order)
	result.Push(slices.Clone(values)...)
	result.version = h.version + 1
	*h = *result
	return nil
}
