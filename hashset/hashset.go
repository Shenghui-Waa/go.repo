// Package hashset implements a set using Go's built-in map.
package hashset

import (
	"iter"

	"github.com/Shenghui-Waa/go.repo/hashmap"
)

type HashSet[T comparable] struct {
	values *hashmap.HashMap[T, struct{}]
}

func New[T comparable](values ...T) *HashSet[T] {
	set := &HashSet[T]{values: hashmap.New[T, struct{}]()}
	for _, value := range values {
		set.Add(value)
	}
	return set
}

func (set *HashSet[T]) Len() int {
	if set.values == nil {
		return 0
	}
	return set.values.Len()
}
func (set *HashSet[T]) IsEmpty() bool { return set.Len() == 0 }
func (set *HashSet[T]) Contains(value T) bool {
	return set.values != nil && set.values.ContainsKey(value)
}

func (set *HashSet[T]) Add(value T) bool {
	if set.values == nil {
		set.values = hashmap.New[T, struct{}]()
	}
	_, existed := set.values.Put(value, struct{}{})
	return !existed
}

func (set *HashSet[T]) Remove(value T) bool {
	if set.values == nil {
		return false
	}
	_, existed := set.values.Remove(value)
	return existed
}

func (set *HashSet[T]) Clear() {
	if set.values != nil {
		set.values.Clear()
	}
}

func (set *HashSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		if set.values != nil {
			set.values.Keys()(yield)
		}
	}
}
