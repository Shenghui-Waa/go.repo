// Package treeset orders and deduplicates members using the local tree map.
package treeset

import (
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/sets"
	"github.com/Shenghui-Waa/go.repo/sets/internal/setcore"
)

type Set[T any] struct{ setcore.Core[T, setcore.Tree] }
type Iterator[T any] = setcore.Iterator[T]

func New[T compare.Comparable[T]]() *Set[T] {
	return NewWithComparator(compare.FromComparable[T]())
}
func NewWithComparator[T any](order compare.Comparator[T]) *Set[T] {
	return &Set[T]{setcore.New[T, setcore.Tree](order)}
}
func (s *Set[T]) Min() (T, bool)                     { var value T; return s.Bound(value, "min") }
func (s *Set[T]) Max() (T, bool)                     { var value T; return s.Bound(value, "max") }
func (s *Set[T]) Floor(value T) (T, bool)            { return s.Bound(value, "floor") }
func (s *Set[T]) Ceiling(value T) (T, bool)          { return s.Bound(value, "ceiling") }
func (s *Set[T]) Map(f func(int, T) T) *Set[T]       { return &Set[T]{s.Core.Map(f)} }
func (s *Set[T]) Select(f func(int, T) bool) *Set[T] { return &Set[T]{s.Core.Select(f)} }
func (s *Set[T]) Union(other sets.Set[T]) *Set[T] {
	return &Set[T]{s.Combine(other, "union")}
}
func (s *Set[T]) Intersection(other sets.Set[T]) *Set[T] {
	return &Set[T]{s.Combine(other, "intersection")}
}
func (s *Set[T]) Difference(other sets.Set[T]) *Set[T] {
	return &Set[T]{s.Combine(other, "difference")}
}
func (s *Set[T]) SymmetricDifference(other sets.Set[T]) *Set[T] {
	return &Set[T]{s.Combine(other, "symmetric")}
}
