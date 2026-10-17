// Package treebidimap maintains two red-black indexes for a one-to-one relation.
package treebidimap

import (
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/maps/internal/mapcore"
)

type Map[K, V any] struct {
	mapcore.Core[K, V, mapcore.TreeBidi]
}
type Iterator[K, V any] = mapcore.Cursor[K, V]

func New[K compare.Comparable[K], V compare.Comparable[V]]() *Map[K, V] {
	return NewWithComparator(compare.FromComparable[K](), compare.FromComparable[V]())
}
func NewWithComparator[K, V any](keys compare.Comparator[K], values compare.Comparator[V]) *Map[K, V] {
	m := &Map[K, V]{}
	m.Init(keys, values)
	return m
}
func (m *Map[K, V]) GetKey(value V) (K, bool) { return m.LookupValue(value) }
func (m *Map[K, V]) RemoveValue(value V) bool {
	key, ok := m.GetKey(value)
	return ok && m.Remove(key)
}
func (m *Map[K, V]) Min() (K, V, bool)          { var key K; return m.Bound(key, "min") }
func (m *Map[K, V]) Max() (K, V, bool)          { var key K; return m.Bound(key, "max") }
func (m *Map[K, V]) Floor(key K) (K, V, bool)   { return m.Bound(key, "floor") }
func (m *Map[K, V]) Ceiling(key K) (K, V, bool) { return m.Bound(key, "ceiling") }
func (m *Map[K, V]) Map(f func(K, V) (K, V)) *Map[K, V] {
	return &Map[K, V]{m.Core.Map(f)}
}
func (m *Map[K, V]) Select(f func(K, V) bool) *Map[K, V] {
	return &Map[K, V]{m.Core.Select(f)}
}
