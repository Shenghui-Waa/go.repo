// Package treeset implements a comparator-ordered set on top of TreeMap.
package treeset

import (
	"iter"

	"github.com/Shenghui-Waa/go.repo/collections"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/treemap"
)

type TreeSet[T any] struct{ values *treemap.TreeMap[T, struct{}] }

func New[T any](less compare.Comparator[T], values ...T) *TreeSet[T] {
	set := &TreeSet[T]{values: treemap.New[T, struct{}](less)}
	for _, value := range values {
		set.Add(value)
	}
	return set
}

func (set *TreeSet[T]) Len() int                          { return set.values.Len() }
func (set *TreeSet[T]) IsEmpty() bool                     { return set.Len() == 0 }
func (set *TreeSet[T]) Comparator() compare.Comparator[T] { return set.values.Comparator() }
func (set *TreeSet[T]) Contains(value T) bool             { return set.values.ContainsKey(value) }
func (set *TreeSet[T]) Add(value T) bool {
	_, existed := set.values.Put(value, struct{}{})
	return !existed
}
func (set *TreeSet[T]) Remove(value T) bool { _, existed := set.values.Remove(value); return existed }
func (set *TreeSet[T]) Clear()              { set.values.Clear() }
func (set *TreeSet[T]) All() iter.Seq[T]    { return set.values.Keys() }
func (set *TreeSet[T]) Backward() iter.Seq[T] {
	return func(yield func(T) bool) { set.values.Backward()(func(key T, _ struct{}) bool { return yield(key) }) }
}
func (set *TreeSet[T]) First() (T, bool) { return set.values.FirstKey() }
func (set *TreeSet[T]) Last() (T, bool)  { return set.values.LastKey() }
func (set *TreeSet[T]) Lower(value T) (T, bool) {
	key, _, ok := set.values.LowerEntry(value)
	return key, ok
}
func (set *TreeSet[T]) Floor(value T) (T, bool) {
	key, _, ok := set.values.FloorEntry(value)
	return key, ok
}
func (set *TreeSet[T]) Ceiling(value T) (T, bool) {
	key, _, ok := set.values.CeilingEntry(value)
	return key, ok
}
func (set *TreeSet[T]) Higher(value T) (T, bool) {
	key, _, ok := set.values.HigherEntry(value)
	return key, ok
}
func (set *TreeSet[T]) Range(from T, fromInclusive bool, to T, toInclusive bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		set.values.Range(from, fromInclusive, to, toInclusive)(func(key T, _ struct{}) bool { return yield(key) })
	}
}

var _ collections.NavigableSet[int] = (*TreeSet[int])(nil)
