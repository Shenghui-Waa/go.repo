package linkedlist

import (
	"github.com/Shenghui-Waa/go.repo/collections"
	"github.com/Shenghui-Waa/go.repo/iterable"
)

func (list *LinkedList[T]) Positions(index int) collections.Position[T] {
	if index < 0 || index > list.size {
		panic("linkedlist: invalid position index")
	}
	return &position[T]{list: list, next: list.nodeAt(index), index: index, version: list.version}
}

type position[T any] struct {
	list    *LinkedList[T]
	next    *node[T]
	last    *node[T]
	index   int
	version uint64
}

func (cursor *position[T]) HasNext() bool     { return cursor.index < cursor.list.size }
func (cursor *position[T]) HasPrevious() bool { return cursor.index > 0 }
func (cursor *position[T]) Next() (T, error) {
	if cursor.version != cursor.list.version {
		var zero T
		return zero, iterable.ErrConcurrentModification
	}
	if cursor.next == nil {
		var zero T
		return zero, iterable.ErrNoSuchElement
	}
	cursor.last = cursor.next
	cursor.next = cursor.next.next
	cursor.index++
	return cursor.last.value, nil
}
func (cursor *position[T]) Previous() (T, error) {
	if cursor.version != cursor.list.version {
		var zero T
		return zero, iterable.ErrConcurrentModification
	}
	if !cursor.HasPrevious() {
		var zero T
		return zero, iterable.ErrNoSuchElement
	}
	if cursor.next == nil {
		cursor.next = cursor.list.tail
	} else {
		cursor.next = cursor.next.prev
	}
	cursor.last = cursor.next
	cursor.index--
	return cursor.last.value, nil
}
func (cursor *position[T]) Add(value T) error {
	if cursor.version != cursor.list.version {
		return iterable.ErrConcurrentModification
	}
	if cursor.next == nil {
		cursor.list.AddLast(value)
	} else {
		item := &node[T]{value: value, prev: cursor.next.prev, next: cursor.next}
		if item.prev == nil {
			cursor.list.head = item
		} else {
			item.prev.next = item
		}
		cursor.next.prev = item
		cursor.list.size++
		cursor.list.version++
	}
	cursor.index++
	cursor.last = nil
	cursor.version = cursor.list.version
	return nil
}
func (cursor *position[T]) Set(value T) error {
	if cursor.version != cursor.list.version {
		return iterable.ErrConcurrentModification
	}
	if cursor.last == nil {
		return iterable.ErrIllegalState
	}
	cursor.last.value = value
	return nil
}
func (cursor *position[T]) Remove() error {
	if cursor.version != cursor.list.version {
		return iterable.ErrConcurrentModification
	}
	if cursor.last == nil {
		return iterable.ErrIllegalState
	}
	if cursor.next == cursor.last {
		cursor.next = cursor.last.next
	} else {
		cursor.index--
	}
	cursor.list.unlink(cursor.last)
	cursor.last = nil
	cursor.version = cursor.list.version
	return nil
}

var _ collections.PositionedList[int] = (*LinkedList[int])(nil)
