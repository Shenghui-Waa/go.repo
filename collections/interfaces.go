// Package collections contains Java Collections-inspired Go contracts.
package collections

import (
	"iter"

	"github.com/Shenghui-Waa/go.repo/compare"
)

type Collection[T any] interface {
	All() iter.Seq[T]
	Len() int
	IsEmpty() bool
	Contains(value T) bool
	Add(value T) bool
	Remove(value T) bool
	Clear()
}

type SequencedCollection[T any] interface {
	Collection[T]
	First() (T, bool)
	Last() (T, bool)
	AddFirst(value T)
	AddLast(value T)
	RemoveFirst() (T, bool)
	RemoveLast() (T, bool)
	Backward() iter.Seq[T]
}

type List[T any] interface {
	SequencedCollection[T]
	Get(index int) (T, bool)
	Set(index int, value T) (T, bool)
	Insert(index int, value T) bool
	RemoveAt(index int) (T, bool)
	IndexOf(value T) int
	LastIndexOf(value T) int
}

type Queue[T any] interface {
	Collection[T]
	Offer(value T) bool
	Peek() (T, bool)
	Poll() (T, bool)
}

type Deque[T any] interface {
	Queue[T]
	SequencedCollection[T]
	OfferFirst(value T) bool
	OfferLast(value T) bool
	AddFirst(value T)
	AddLast(value T)
	PeekFirst() (T, bool)
	PeekLast() (T, bool)
	PollFirst() (T, bool)
	PollLast() (T, bool)
}

type Set[T any] interface {
	Collection[T]
}

type Map[K any, V any] interface {
	All() iter.Seq2[K, V]
	Keys() iter.Seq[K]
	Values() iter.Seq[V]
	Get(key K) (V, bool)
	Put(key K, value V) (V, bool)
	Remove(key K) (V, bool)
	ContainsKey(key K) bool
	Len() int
	Clear()
}

type SortedSet[T any] interface {
	Set[T]
	Comparator() compare.Comparator[T]
	First() (T, bool)
	Last() (T, bool)
	Backward() iter.Seq[T]
}

type NavigableSet[T any] interface {
	SortedSet[T]
	Lower(value T) (T, bool)
	Floor(value T) (T, bool)
	Ceiling(value T) (T, bool)
	Higher(value T) (T, bool)
	Range(from T, fromInclusive bool, to T, toInclusive bool) iter.Seq[T]
}

type SortedMap[K any, V any] interface {
	Map[K, V]
	Comparator() compare.Comparator[K]
	FirstKey() (K, bool)
	LastKey() (K, bool)
	Backward() iter.Seq2[K, V]
}

type NavigableMap[K any, V any] interface {
	SortedMap[K, V]
	LowerEntry(key K) (entryKey K, value V, ok bool)
	FloorEntry(key K) (entryKey K, value V, ok bool)
	CeilingEntry(key K) (entryKey K, value V, ok bool)
	HigherEntry(key K) (entryKey K, value V, ok bool)
	Range(from K, fromInclusive bool, to K, toInclusive bool) iter.Seq2[K, V]
}

type Position[T any] interface {
	HasNext() bool
	Next() (T, error)
	HasPrevious() bool
	Previous() (T, error)
	Add(value T) error
	Set(value T) error
	Remove() error
}

type PositionedList[T any] interface {
	List[T]
	Positions(index int) Position[T]
}
