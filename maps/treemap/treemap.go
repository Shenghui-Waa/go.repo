// Package treemap maintains keys in comparator order with a red-black tree.
package treemap

import (
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/maps/internal/mapcore"
)

type Map[K, V any] struct {
	mapcore.Core[K, V, mapcore.Tree]
}
type Iterator[K, V any] = mapcore.Cursor[K, V]

// New requires natural key ordering; built-in keys use NewWithComparator.
func New[K compare.Comparable[K], V any]() *Map[K, V] {
	return NewWithComparator[K, V](compare.FromComparable[K]())
}
func NewWithComparator[K, V any](order compare.Comparator[K]) *Map[K, V] {
	m := &Map[K, V]{}
	m.Init(order, nil)
	return m
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
