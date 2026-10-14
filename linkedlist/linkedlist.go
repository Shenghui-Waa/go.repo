// Package linkedlist implements a doubly linked list and deque.
package linkedlist

import (
	"iter"

	"github.com/Shenghui-Waa/go.repo/collections"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/iterable"
)

type node[T any] struct {
	value T
	prev  *node[T]
	next  *node[T]
}

type LinkedList[T any] struct {
	head    *node[T]
	tail    *node[T]
	size    int
	equal   compare.Equal[T]
	version uint64
}

func New[T any](equal compare.Equal[T], values ...T) *LinkedList[T] {
	if equal == nil {
		panic("linkedlist: nil equality strategy")
	}
	list := &LinkedList[T]{equal: equal}
	for _, value := range values {
		list.AddLast(value)
	}
	return list
}

func NewComparable[T comparable](values ...T) *LinkedList[T] {
	return New(compare.NativeEqual[T], values...)
}

func (list *LinkedList[T]) Len() int {
	return list.size
}

func (list *LinkedList[T]) IsEmpty() bool {
	return list.size == 0
}

func (list *LinkedList[T]) Add(value T) bool {
	list.AddLast(value)
	return true
}

func (list *LinkedList[T]) AddFirst(value T) {
	item := &node[T]{value: value, next: list.head}
	if list.head == nil {
		list.tail = item
	} else {
		list.head.prev = item
	}
	list.head = item
	list.size++
	list.version++
}

func (list *LinkedList[T]) AddLast(value T) {
	item := &node[T]{value: value, prev: list.tail}
	if list.tail == nil {
		list.head = item
	} else {
		list.tail.next = item
	}
	list.tail = item
	list.size++
	list.version++
}

func (list *LinkedList[T]) Offer(value T) bool {
	return list.OfferLast(value)
}

func (list *LinkedList[T]) OfferFirst(value T) bool {
	list.AddFirst(value)
	return true
}

func (list *LinkedList[T]) OfferLast(value T) bool {
	list.AddLast(value)
	return true
}

func (list *LinkedList[T]) First() (T, bool) {
	return list.PeekFirst()
}

func (list *LinkedList[T]) Last() (T, bool) {
	return list.PeekLast()
}

func (list *LinkedList[T]) Peek() (T, bool) {
	return list.PeekFirst()
}

func (list *LinkedList[T]) PeekFirst() (T, bool) {
	if list.head == nil {
		var zero T
		return zero, false
	}
	return list.head.value, true
}

func (list *LinkedList[T]) PeekLast() (T, bool) {
	if list.tail == nil {
		var zero T
		return zero, false
	}
	return list.tail.value, true
}

func (list *LinkedList[T]) Poll() (T, bool) {
	return list.PollFirst()
}

func (list *LinkedList[T]) PollFirst() (T, bool) {
	return list.RemoveFirst()
}

func (list *LinkedList[T]) PollLast() (T, bool) {
	return list.RemoveLast()
}

func (list *LinkedList[T]) RemoveFirst() (T, bool) {
	return list.unlink(list.head)
}

func (list *LinkedList[T]) RemoveLast() (T, bool) {
	return list.unlink(list.tail)
}

func (list *LinkedList[T]) Contains(value T) bool {
	return list.IndexOf(value) >= 0
}

func (list *LinkedList[T]) IndexOf(value T) int {
	index := 0
	for current := list.head; current != nil; current = current.next {
		if list.equal(current.value, value) {
			return index
		}
		index++
	}
	return -1
}

func (list *LinkedList[T]) LastIndexOf(value T) int {
	index := list.size - 1
	for current := list.tail; current != nil; current = current.prev {
		if list.equal(current.value, value) {
			return index
		}
		index--
	}
	return -1
}

func (list *LinkedList[T]) Get(index int) (T, bool) {
	item := list.nodeAt(index)
	if item == nil {
		var zero T
		return zero, false
	}
	return item.value, true
}

func (list *LinkedList[T]) Set(index int, value T) (T, bool) {
	item := list.nodeAt(index)
	if item == nil {
		var zero T
		return zero, false
	}
	old := item.value
	item.value = value
	return old, true
}

func (list *LinkedList[T]) Insert(index int, value T) bool {
	if index < 0 || index > list.size {
		return false
	}
	if index == 0 {
		list.AddFirst(value)
		return true
	}
	if index == list.size {
		list.AddLast(value)
		return true
	}
	next := list.nodeAt(index)
	item := &node[T]{value: value, prev: next.prev, next: next}
	next.prev.next = item
	next.prev = item
	list.size++
	list.version++
	return true
}

func (list *LinkedList[T]) Remove(value T) bool {
	for current := list.head; current != nil; current = current.next {
		if list.equal(current.value, value) {
			_, _ = list.unlink(current)
			return true
		}
	}
	return false
}

func (list *LinkedList[T]) RemoveAt(index int) (T, bool) {
	return list.unlink(list.nodeAt(index))
}

func (list *LinkedList[T]) Clear() {
	if list.size == 0 {
		return
	}
	for current := list.head; current != nil; {
		next := current.next
		current.prev = nil
		current.next = nil
		current = next
	}
	list.head = nil
	list.tail = nil
	list.size = 0
	list.version++
}

func (list *LinkedList[T]) All() iter.Seq[T] {
	expectedVersion := list.version
	return func(yield func(T) bool) {
		for current := list.head; current != nil; current = current.next {
			if expectedVersion != list.version {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(current.value) {
				return
			}
		}
	}
}

func (list *LinkedList[T]) Backward() iter.Seq[T] {
	expectedVersion := list.version
	return func(yield func(T) bool) {
		for current := list.tail; current != nil; current = current.prev {
			if expectedVersion != list.version {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(current.value) {
				return
			}
		}
	}
}

func (list *LinkedList[T]) nodeAt(index int) *node[T] {
	if index < 0 || index >= list.size {
		return nil
	}
	if index < list.size/2 {
		current := list.head
		for step := 0; step < index; step++ {
			current = current.next
		}
		return current
	}
	current := list.tail
	for step := list.size - 1; step > index; step-- {
		current = current.prev
	}
	return current
}

func (list *LinkedList[T]) unlink(item *node[T]) (T, bool) {
	if item == nil {
		var zero T
		return zero, false
	}
	if item.prev == nil {
		list.head = item.next
	} else {
		item.prev.next = item.next
	}
	if item.next == nil {
		list.tail = item.prev
	} else {
		item.next.prev = item.prev
	}
	value := item.value
	item.prev = nil
	item.next = nil
	list.size--
	list.version++
	return value, true
}

var _ collections.List[int] = (*LinkedList[int])(nil)
var _ collections.Deque[int] = (*LinkedList[int])(nil)
