package java

import (
	"errors"
	"iter"

	"github.com/Shenghui-Waa/go.repo/collections"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/iterable"
)

var ErrIndexOutOfBounds = errors.New("index out of bounds")

type Map[K any, V any] struct {
	Core  collections.Map[K, V]
	Equal compare.Equal[V]
}

func NewMap[K any, V any](core collections.Map[K, V], equal compare.Equal[V]) Map[K, V] {
	if core == nil {
		panic("java: nil map")
	}
	if equal == nil {
		panic("java: nil value equality strategy")
	}
	return Map[K, V]{Core: core, Equal: equal}
}

func (mapping Map[K, V]) Size() int                    { return mapping.Core.Len() }
func (mapping Map[K, V]) Get(key K) (V, bool)          { return mapping.Core.Get(key) }
func (mapping Map[K, V]) Put(key K, value V) (V, bool) { return mapping.Core.Put(key, value) }
func (mapping Map[K, V]) Remove(key K) (V, bool)       { return mapping.Core.Remove(key) }
func (mapping Map[K, V]) KeySet() KeySet[K, V]         { return KeySet[K, V]{core: mapping.Core} }
func (mapping Map[K, V]) Values() Values[K, V] {
	return Values[K, V]{core: mapping.Core, equal: mapping.Equal}
}
func (mapping Map[K, V]) EntrySet() EntrySet[K, V] {
	return EntrySet[K, V]{core: mapping.Core, equal: mapping.Equal}
}

type KeySet[K any, V any] struct{ core collections.Map[K, V] }

func (view KeySet[K, V]) Len() int            { return view.core.Len() }
func (view KeySet[K, V]) IsEmpty() bool       { return view.Len() == 0 }
func (view KeySet[K, V]) Contains(key K) bool { return view.core.ContainsKey(key) }
func (view KeySet[K, V]) Remove(key K) bool   { _, ok := view.core.Remove(key); return ok }
func (view KeySet[K, V]) Clear()              { view.core.Clear() }
func (view KeySet[K, V]) Add(key K) bool      { panic(iterable.ErrUnsupportedOperation) }
func (view KeySet[K, V]) All() iter.Seq[K]    { return view.core.Keys() }

type Values[K any, V any] struct {
	core  collections.Map[K, V]
	equal compare.Equal[V]
}

func (view Values[K, V]) Len() int         { return view.core.Len() }
func (view Values[K, V]) IsEmpty() bool    { return view.Len() == 0 }
func (view Values[K, V]) Clear()           { view.core.Clear() }
func (view Values[K, V]) Add(value V) bool { panic(iterable.ErrUnsupportedOperation) }
func (view Values[K, V]) All() iter.Seq[V] { return view.core.Values() }
func (view Values[K, V]) Contains(value V) bool {
	for current := range view.All() {
		if view.equal(current, value) {
			return true
		}
	}
	return false
}
func (view Values[K, V]) Remove(value V) bool {
	for key, current := range view.core.All() {
		if view.equal(current, value) {
			view.core.Remove(key)
			return true
		}
	}
	return false
}

// Entry is a live key-bound view; SetValue writes through to its source map.
type Entry[K any, V any] struct {
	Key  K
	core collections.Map[K, V]
}

func (entry Entry[K, V]) Value() (V, bool) { return entry.core.Get(entry.Key) }
func (entry Entry[K, V]) SetValue(value V) (V, error) {
	if !entry.core.ContainsKey(entry.Key) {
		var zero V
		return zero, iterable.ErrIllegalState
	}
	old, _ := entry.core.Put(entry.Key, value)
	return old, nil
}

type EntrySet[K any, V any] struct {
	core  collections.Map[K, V]
	equal compare.Equal[V]
}

func (view EntrySet[K, V]) Len() int { return view.core.Len() }
func (view EntrySet[K, V]) Clear()   { view.core.Clear() }
func (view EntrySet[K, V]) All() iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		for key := range view.core.Keys() {
			if !yield(Entry[K, V]{Key: key, core: view.core}) {
				return
			}
		}
	}
}
func (view EntrySet[K, V]) Remove(key K, value V) bool {
	current, ok := view.core.Get(key)
	if !ok || !view.equal(current, value) {
		return false
	}
	view.core.Remove(key)
	return true
}
