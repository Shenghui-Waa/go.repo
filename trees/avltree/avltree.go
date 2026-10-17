// Package avltree implements a locally maintained AVL search tree.
package avltree

import (
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/trees/internal/ordered"
)

type Tree[K, V any] struct {
	ordered.Core[K, V, ordered.AVL]
}
type Iterator[K, V any] = ordered.Iterator[K, V]

func New[K compare.Comparable[K], V any]() *Tree[K, V] {
	return NewWithComparator[K, V](compare.FromComparable[K]())
}
func NewWithComparator[K, V any](order compare.Comparator[K]) *Tree[K, V] {
	return &Tree[K, V]{ordered.New[K, V, ordered.AVL](order, 0)}
}
func (t *Tree[K, V]) Map(f func(K, V) (K, V)) *Tree[K, V] {
	return &Tree[K, V]{t.Core.Map(f)}
}
func (t *Tree[K, V]) Select(f func(K, V) bool) *Tree[K, V] {
	return &Tree[K, V]{t.Core.Select(f)}
}
