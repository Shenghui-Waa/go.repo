// Package btree implements a local B-tree with configurable maximum children.
package btree

import (
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/trees/internal/ordered"
)

const DefaultOrder = 32

type Tree[K, V any] struct {
	ordered.Core[K, V, ordered.BTree]
}
type Iterator[K, V any] = ordered.Iterator[K, V]

func New[K compare.Comparable[K], V any]() *Tree[K, V] {
	return NewWithComparator[K, V](compare.FromComparable[K]())
}
func NewWithComparator[K, V any](order compare.Comparator[K]) *Tree[K, V] {
	return NewWithOrderAndComparator[K, V](DefaultOrder, order)
}
func NewWithOrder[K compare.Comparable[K], V any](degree int) *Tree[K, V] {
	return NewWithOrderAndComparator[K, V](degree, compare.FromComparable[K]())
}
func NewWithOrderAndComparator[K, V any](degree int, order compare.Comparator[K]) *Tree[K, V] {
	return &Tree[K, V]{ordered.New[K, V, ordered.BTree](order, degree)}
}
func (t *Tree[K, V]) Map(f func(K, V) (K, V)) *Tree[K, V] {
	return &Tree[K, V]{t.Core.Map(f)}
}
func (t *Tree[K, V]) Select(f func(K, V) bool) *Tree[K, V] {
	return &Tree[K, V]{t.Core.Select(f)}
}
