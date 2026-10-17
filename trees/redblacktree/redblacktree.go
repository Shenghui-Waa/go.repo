// Package redblacktree implements a locally maintained red-black search tree.
package redblacktree

import (
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/trees/internal/ordered"
)

type Tree[K, V any] struct {
	ordered.Core[K, V, ordered.RedBlack]
}
type Iterator[K, V any] = ordered.Iterator[K, V]

func New[K compare.Comparable[K], V any]() *Tree[K, V] {
	return NewWithComparator[K, V](compare.FromComparable[K]())
}
func NewWithComparator[K, V any](order compare.Comparator[K]) *Tree[K, V] {
	return &Tree[K, V]{ordered.New[K, V, ordered.RedBlack](order, 0)}
}
func (t *Tree[K, V]) Map(f func(K, V) (K, V)) *Tree[K, V] {
	return &Tree[K, V]{t.Core.Map(f)}
}
func (t *Tree[K, V]) Select(f func(K, V) bool) *Tree[K, V] {
	return &Tree[K, V]{t.Core.Select(f)}
}
