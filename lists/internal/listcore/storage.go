package listcore

import "slices"

// Storage separates shared contracts from actual array and node algorithms.
type storage[T any] interface {
	len() int
	get(int) T
	set(int, T)
	insert(int, []T)
	remove(int)
	clear()
	reader() func(int) T
}

type array[T any] struct{ values []T }

func (a *array[T]) len() int           { return len(a.values) }
func (a *array[T]) get(i int) T        { return a.values[i] }
func (a *array[T]) set(i int, value T) { a.values[i] = value }
func (a *array[T]) insert(i int, values []T) {
	a.values = slices.Insert(a.values, i, values...)
}
func (a *array[T]) remove(i int)        { a.values = slices.Delete(a.values, i, i+1) }
func (a *array[T]) clear()              { a.values = nil }
func (a *array[T]) reader() func(int) T { return a.get }

type singleNode[T any] struct {
	value T
	next  *singleNode[T]
}

type single[T any] struct {
	head, tail *singleNode[T]
	size       int
}

func (s *single[T]) len() int { return s.size }
func (s *single[T]) node(i int) *singleNode[T] {
	if i == s.size-1 {
		return s.tail
	}
	n := s.head
	for range i {
		n = n.next
	}
	return n
}
func (s *single[T]) get(i int) T        { return s.node(i).value }
func (s *single[T]) set(i int, value T) { s.node(i).value = value }
func (s *single[T]) insert(i int, values []T) {
	if len(values) == 0 {
		return
	}
	var first, last *singleNode[T]
	for _, value := range values {
		n := &singleNode[T]{value: value}
		if first == nil {
			first = n
		} else {
			last.next = n
		}
		last = n
	}
	if i == 0 {
		last.next = s.head
		s.head = first
	} else {
		previous := s.node(i - 1)
		last.next = previous.next
		previous.next = first
	}
	if i == s.size {
		s.tail = last
	}
	s.size += len(values)
}
func (s *single[T]) remove(i int) {
	if i == 0 {
		s.head = s.head.next
		if s.size == 1 {
			s.tail = nil
		}
	} else {
		previous := s.node(i - 1)
		previous.next = previous.next.next
		if i == s.size-1 {
			s.tail = previous
		}
	}
	s.size--
}
func (s *single[T]) clear() { s.head, s.tail, s.size = nil, nil, 0 }
func (s *single[T]) reader() func(int) T {
	var current *singleNode[T]
	index := -1
	return func(i int) T {
		if current != nil && i == index+1 {
			current = current.next
		} else if current == nil || i != index {
			current = s.node(i)
		}
		index = i
		return current.value
	}
}

type doubleNode[T any] struct {
	value          T
	next, previous *doubleNode[T]
}

type double[T any] struct {
	head, tail *doubleNode[T]
	size       int
}

func (d *double[T]) len() int { return d.size }
func (d *double[T]) node(i int) *doubleNode[T] {
	if i < d.size/2 {
		n := d.head
		for range i {
			n = n.next
		}
		return n
	}
	n := d.tail
	for j := d.size - 1; j > i; j-- {
		n = n.previous
	}
	return n
}
func (d *double[T]) get(i int) T        { return d.node(i).value }
func (d *double[T]) set(i int, value T) { d.node(i).value = value }
func (d *double[T]) insert(i int, values []T) {
	if len(values) == 0 {
		return
	}
	var previous, next *doubleNode[T]
	if i == d.size {
		previous = d.tail
	} else {
		next = d.node(i)
		previous = next.previous
	}
	for _, value := range values {
		n := &doubleNode[T]{value: value, previous: previous}
		if previous == nil {
			d.head = n
		} else {
			previous.next = n
		}
		previous = n
	}
	previous.next = next
	if next == nil {
		d.tail = previous
	} else {
		next.previous = previous
	}
	d.size += len(values)
}
func (d *double[T]) remove(i int) {
	n := d.node(i)
	if n.previous == nil {
		d.head = n.next
	} else {
		n.previous.next = n.next
	}
	if n.next == nil {
		d.tail = n.previous
	} else {
		n.next.previous = n.previous
	}
	d.size--
}
func (d *double[T]) clear() { d.head, d.tail, d.size = nil, nil, 0 }
func (d *double[T]) reader() func(int) T {
	var current *doubleNode[T]
	index := -1
	return func(i int) T {
		switch {
		case current != nil && i == index+1:
			current = current.next
		case current != nil && i == index-1:
			current = current.previous
		case current == nil || i != index:
			current = d.node(i)
		}
		index = i
		return current.value
	}
}
