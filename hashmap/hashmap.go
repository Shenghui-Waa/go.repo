// Package hashmap wraps Go's built-in map with collection-style operations.
package hashmap

import (
	"iter"

	"github.com/Shenghui-Waa/go.repo/iterable"
)

type HashMap[K comparable, V any] struct {
	values  map[K]V
	version uint64
}

func New[K comparable, V any]() *HashMap[K, V] {
	return &HashMap[K, V]{values: make(map[K]V)}
}

func (mapping *HashMap[K, V]) Len() int { return len(mapping.values) }

func (mapping *HashMap[K, V]) IsEmpty() bool { return len(mapping.values) == 0 }

func (mapping *HashMap[K, V]) Get(key K) (V, bool) {
	value, ok := mapping.values[key]
	return value, ok
}

func (mapping *HashMap[K, V]) Put(key K, value V) (V, bool) {
	if mapping.values == nil {
		mapping.values = make(map[K]V)
	}
	old, replaced := mapping.values[key]
	mapping.values[key] = value
	if !replaced {
		mapping.version++
	}
	return old, replaced
}

func (mapping *HashMap[K, V]) Remove(key K) (V, bool) {
	value, ok := mapping.values[key]
	if ok {
		delete(mapping.values, key)
		mapping.version++
	}
	return value, ok
}

func (mapping *HashMap[K, V]) ContainsKey(key K) bool {
	_, ok := mapping.values[key]
	return ok
}

func (mapping *HashMap[K, V]) Clear() {
	if len(mapping.values) == 0 {
		return
	}
	clear(mapping.values)
	mapping.version++
}

func (mapping *HashMap[K, V]) All() iter.Seq2[K, V] {
	expectedVersion := mapping.version
	return func(yield func(K, V) bool) {
		for key, value := range mapping.values {
			if expectedVersion != mapping.version {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(key, value) {
				return
			}
		}
	}
}

func (mapping *HashMap[K, V]) Keys() iter.Seq[K] {
	expectedVersion := mapping.version
	return func(yield func(K) bool) {
		for key := range mapping.values {
			if expectedVersion != mapping.version {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(key) {
				return
			}
		}
	}
}

func (mapping *HashMap[K, V]) Values() iter.Seq[V] {
	expectedVersion := mapping.version
	return func(yield func(V) bool) {
		for _, value := range mapping.values {
			if expectedVersion != mapping.version {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(value) {
				return
			}
		}
	}
}
