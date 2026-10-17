package queuecore

import "github.com/Shenghui-Waa/go.repo/trees/binaryheap"

type storage[T any] interface {
	push(T)
	pop() (T, bool)
	peek() (T, bool)
	len() int
	clear()
	values() []T
}

// ring is either dynamically growing or fixed-capacity and overwriting.
type ring[T any] struct {
	buffer []T
	head   int
	size   int
	fixed  bool
}

func (r *ring[T]) len() int { return r.size }
func (r *ring[T]) push(value T) {
	if r.size == len(r.buffer) {
		if r.fixed {
			r.buffer[r.head] = value
			r.head = (r.head + 1) % len(r.buffer)
			return
		}
		grown := make([]T, max(DefaultCapacity, 2*len(r.buffer)))
		for i := 0; i < r.size; i++ {
			grown[i] = r.buffer[(r.head+i)%len(r.buffer)]
		}
		r.buffer, r.head = grown, 0
	}
	r.buffer[(r.head+r.size)%len(r.buffer)] = value
	r.size++
}
func (r *ring[T]) pop() (T, bool) {
	var zero T
	if r.size == 0 {
		return zero, false
	}
	value := r.buffer[r.head]
	r.buffer[r.head] = zero
	r.head = (r.head + 1) % len(r.buffer)
	r.size--
	if r.size == 0 {
		r.head = 0
	}
	return value, true
}
func (r *ring[T]) peek() (T, bool) {
	if r.size == 0 {
		var zero T
		return zero, false
	}
	return r.buffer[r.head], true
}
func (r *ring[T]) clear() {
	if r.fixed {
		clear(r.buffer)
	} else {
		r.buffer = nil
	}
	r.head, r.size = 0, 0
}
func (r *ring[T]) values() []T {
	result := make([]T, r.size)
	for i := range result {
		result[i] = r.buffer[(r.head+i)%len(r.buffer)]
	}
	return result
}

type node[T any] struct {
	value T
	next  *node[T]
}
type linked[T any] struct {
	head, tail *node[T]
	size       int
}

func (l *linked[T]) len() int { return l.size }
func (l *linked[T]) push(value T) {
	n := &node[T]{value: value}
	if l.tail == nil {
		l.head = n
	} else {
		l.tail.next = n
	}
	l.tail = n
	l.size++
}
func (l *linked[T]) pop() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	}
	value := l.head.value
	l.head = l.head.next
	l.size--
	if l.head == nil {
		l.tail = nil
	}
	return value, true
}
func (l *linked[T]) peek() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	}
	return l.head.value, true
}
func (l *linked[T]) clear() { l.head, l.tail, l.size = nil, nil, 0 }
func (l *linked[T]) values() []T {
	result := make([]T, 0, l.size)
	for n := l.head; n != nil; n = n.next {
		result = append(result, n.value)
	}
	return result
}

// priority reuses the project's local heap without imposing object equality.
type priority[T any] struct{ heap *binaryheap.Heap[T] }

func (p *priority[T]) len() int        { return p.heap.Size() }
func (p *priority[T]) push(value T)    { p.heap.Push(value) }
func (p *priority[T]) pop() (T, bool)  { return p.heap.Pop() }
func (p *priority[T]) peek() (T, bool) { return p.heap.Peek() }
func (p *priority[T]) clear()          { p.heap.Clear() }
func (p *priority[T]) values() []T     { return p.heap.Values() }
