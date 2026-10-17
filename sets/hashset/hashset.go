// Package hashset implements membership using the project's hash map.
package hashset

import (
	"github.com/Shenghui-Waa/go.repo/sets"
	"github.com/Shenghui-Waa/go.repo/sets/internal/setcore"
)

type Set[T any] struct{ setcore.Core[T, setcore.Hash] }
type Iterator[T any] = setcore.Iterator[T]

func New[T any]() *Set[T]                            { return &Set[T]{} }
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
