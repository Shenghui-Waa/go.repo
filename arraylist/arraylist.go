// Package arraylist implements a resizable array-backed list.
package arraylist

import (
	"iter"

	"github.com/Shenghui-Waa/go.repo/collections"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/iterable"
)

type ArrayList[T any] struct {
	values  []T
	equal   compare.Equal[T]
	version uint64
}

func New[T any](equal compare.Equal[T], values ...T) *ArrayList[T] {
	if equal == nil {
		panic("arraylist: nil equality strategy")
	}
	list := &ArrayList[T]{equal: equal}
	list.values = append(list.values, values...)
	return list
}

func NewComparable[T comparable](values ...T) *ArrayList[T] {
	return New(compare.NativeEqual[T], values...)
}

func (list *ArrayList[T]) Len() int {
	return len(list.values)
}

func (list *ArrayList[T]) IsEmpty() bool {
	return len(list.values) == 0
}

func (list *ArrayList[T]) Add(value T) bool {
	list.AddLast(value)
	return true
}

func (list *ArrayList[T]) AddFirst(value T) {
	list.Insert(0, value)
}

func (list *ArrayList[T]) AddLast(value T) {
	list.values = append(list.values, value)
	list.version++
}

func (list *ArrayList[T]) Insert(index int, value T) bool {
	if index < 0 || index > len(list.values) {
		return false
	}
	list.values = append(list.values, value)
	copy(list.values[index+1:], list.values[index:])
	list.values[index] = value
	list.version++
	return true
}

func (list *ArrayList[T]) Get(index int) (T, bool) {
	if index < 0 || index >= len(list.values) {
		var zero T
		return zero, false
	}
	return list.values[index], true
}

func (list *ArrayList[T]) Set(index int, value T) (T, bool) {
	if index < 0 || index >= len(list.values) {
		var zero T
		return zero, false
	}
	old := list.values[index]
	list.values[index] = value
	return old, true
}

func (list *ArrayList[T]) First() (T, bool) {
	return list.Get(0)
}

func (list *ArrayList[T]) Last() (T, bool) {
	return list.Get(len(list.values) - 1)
}

func (list *ArrayList[T]) Contains(value T) bool {
	return list.IndexOf(value) >= 0
}

func (list *ArrayList[T]) IndexOf(value T) int {
	for index, current := range list.values {
		if list.equal(current, value) {
			return index
		}
	}
	return -1
}

func (list *ArrayList[T]) LastIndexOf(value T) int {
	for index := len(list.values) - 1; index >= 0; index-- {
		if list.equal(list.values[index], value) {
			return index
		}
	}
	return -1
}

func (list *ArrayList[T]) Remove(value T) bool {
	index := list.IndexOf(value)
	if index < 0 {
		return false
	}
	_, _ = list.RemoveAt(index)
	return true
}

func (list *ArrayList[T]) RemoveAt(index int) (T, bool) {
	if index < 0 || index >= len(list.values) {
		var zero T
		return zero, false
	}
	old := list.values[index]
	copy(list.values[index:], list.values[index+1:])
	var zero T
	list.values[len(list.values)-1] = zero
	list.values = list.values[:len(list.values)-1]
	list.version++
	return old, true
}

func (list *ArrayList[T]) RemoveFirst() (T, bool) {
	return list.RemoveAt(0)
}

func (list *ArrayList[T]) RemoveLast() (T, bool) {
	return list.RemoveAt(len(list.values) - 1)
}

func (list *ArrayList[T]) Clear() {
	if len(list.values) == 0 {
		return
	}
	clear(list.values)
	list.values = nil
	list.version++
}

func (list *ArrayList[T]) All() iter.Seq[T] {
	expectedVersion := list.version
	return func(yield func(T) bool) {
		for _, value := range list.values {
			if expectedVersion != list.version {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(value) {
				return
			}
		}
	}
}

func (list *ArrayList[T]) Backward() iter.Seq[T] {
	expectedVersion := list.version
	return func(yield func(T) bool) {
		for index := len(list.values) - 1; index >= 0; index-- {
			if expectedVersion != list.version {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(list.values[index]) {
				return
			}
		}
	}
}

func (list *ArrayList[T]) Positions(index int) collections.Position[T] {
	if index < 0 || index > len(list.values) {
		panic("arraylist: invalid position index")
	}
	return &position[T]{
		list:            list,
		nextIndex:       index,
		lastIndex:       -1,
		expectedVersion: list.version,
	}
}

type position[T any] struct {
	list            *ArrayList[T]
	nextIndex       int
	lastIndex       int
	expectedVersion uint64
}

func (position *position[T]) valid() bool {
	return position.expectedVersion == position.list.version
}

func (position *position[T]) HasNext() bool {
	return position.valid() && position.nextIndex < position.list.Len()
}

func (position *position[T]) Next() (T, error) {
	if !position.valid() {
		var zero T
		return zero, iterable.ErrConcurrentModification
	}
	value, ok := position.list.Get(position.nextIndex)
	if !ok {
		var zero T
		return zero, iterable.ErrNoSuchElement
	}
	position.lastIndex = position.nextIndex
	position.nextIndex++
	return value, nil
}

func (position *position[T]) HasPrevious() bool {
	return position.valid() && position.nextIndex > 0
}

func (position *position[T]) Previous() (T, error) {
	if !position.valid() {
		var zero T
		return zero, iterable.ErrConcurrentModification
	}
	if position.nextIndex <= 0 {
		var zero T
		return zero, iterable.ErrNoSuchElement
	}
	position.nextIndex--
	position.lastIndex = position.nextIndex
	value, _ := position.list.Get(position.nextIndex)
	return value, nil
}

func (position *position[T]) Add(value T) error {
	if !position.valid() {
		return iterable.ErrConcurrentModification
	}
	position.list.Insert(position.nextIndex, value)
	position.nextIndex++
	position.lastIndex = -1
	position.expectedVersion = position.list.version
	return nil
}

func (position *position[T]) Set(value T) error {
	if !position.valid() {
		return iterable.ErrConcurrentModification
	}
	if position.lastIndex < 0 {
		return iterable.ErrIllegalState
	}
	_, _ = position.list.Set(position.lastIndex, value)
	return nil
}

func (position *position[T]) Remove() error {
	if !position.valid() {
		return iterable.ErrConcurrentModification
	}
	if position.lastIndex < 0 {
		return iterable.ErrIllegalState
	}
	_, _ = position.list.RemoveAt(position.lastIndex)
	if position.lastIndex < position.nextIndex {
		position.nextIndex--
	}
	position.lastIndex = -1
	position.expectedVersion = position.list.version
	return nil
}

var _ collections.PositionedList[int] = (*ArrayList[int])(nil)
