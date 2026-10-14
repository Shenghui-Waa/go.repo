// Package java provides Java-shaped adapters over Go-native collection APIs.
package java

import (
	"iter"

	"github.com/Shenghui-Waa/go.repo/collections"
	"github.com/Shenghui-Waa/go.repo/iterable"
)

type Queue[T any] struct{ Core collections.Queue[T] }

func (queue Queue[T]) Add(value T) bool { return queue.Core.Offer(value) }

func (queue Queue[T]) Element() (T, error) {
	value, ok := queue.Core.Peek()
	if !ok {
		var zero T
		return zero, iterable.ErrNoSuchElement
	}
	return value, nil
}

func (queue Queue[T]) Remove() (T, error) {
	value, ok := queue.Core.Poll()
	if !ok {
		var zero T
		return zero, iterable.ErrNoSuchElement
	}
	return value, nil
}

func Iterator[T any](source interface{ All() iter.Seq[T] }) iterable.Iterator[T] {
	return iterable.FromSeq(source.All())
}

type Collection[T any] struct{ Core collections.Collection[T] }

func (collection Collection[T]) Size() int                      { return collection.Core.Len() }
func (collection Collection[T]) IsEmpty() bool                  { return collection.Core.IsEmpty() }
func (collection Collection[T]) Add(value T) bool               { return collection.Core.Add(value) }
func (collection Collection[T]) Remove(value T) bool            { return collection.Core.Remove(value) }
func (collection Collection[T]) Contains(value T) bool          { return collection.Core.Contains(value) }
func (collection Collection[T]) Clear()                         { collection.Core.Clear() }
func (collection Collection[T]) Iterator() iterable.Iterator[T] { return Iterator(collection.Core) }

type ListIterator[T any] interface {
	collections.Position[T]
}

type List[T any] struct{ Core collections.PositionedList[T] }

func (list List[T]) Size() int { return list.Core.Len() }
func (list List[T]) Iterator() iterable.Iterator[T] {
	return &positionIterator[T]{Position: list.Core.Positions(0)}
}
func (list List[T]) ListIterator(index int) (ListIterator[T], error) {
	if index < 0 || index > list.Core.Len() {
		return nil, ErrIndexOutOfBounds
	}
	return list.Core.Positions(index), nil
}
func (list List[T]) Get(index int) (T, error) {
	value, ok := list.Core.Get(index)
	if !ok {
		return value, ErrIndexOutOfBounds
	}
	return value, nil
}
func (list List[T]) Set(index int, value T) (T, error) {
	old, ok := list.Core.Set(index, value)
	if !ok {
		return old, ErrIndexOutOfBounds
	}
	return old, nil
}
func (list List[T]) Add(value T) bool { return list.Core.Add(value) }
func (list List[T]) AddAt(index int, value T) error {
	if !list.Core.Insert(index, value) {
		return ErrIndexOutOfBounds
	}
	return nil
}
func (list List[T]) RemoveAt(index int) (T, error) {
	value, ok := list.Core.RemoveAt(index)
	if !ok {
		return value, ErrIndexOutOfBounds
	}
	return value, nil
}

type positionIterator[T any] struct{ collections.Position[T] }

func (iterator *positionIterator[T]) Close() {}
