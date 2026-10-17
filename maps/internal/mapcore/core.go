package mapcore

import (
	"encoding/json"
	"fmt"
	"iter"

	"github.com/Shenghui-Waa/go.repo/base"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/maps"
)

type Kind interface{ flags() (tree, bidi bool) }
type Hash struct{}
type Linked struct{}
type Tree struct{}
type HashBidi struct{}
type TreeBidi struct{}

func (Hash) flags() (bool, bool)     { return false, false }
func (Linked) flags() (bool, bool)   { return false, false }
func (Tree) flags() (bool, bool)     { return true, false }
func (HashBidi) flags() (bool, bool) { return false, true }
func (TreeBidi) flags() (bool, bool) { return true, true }

type entry[K, V any] struct {
	key            K
	value          V
	previous, next *entry[K, V]
}

// Core shares mutation/version rules while keeping indexes private.
type Core[K, V any, M Kind] struct {
	forward    *index[K, *entry[K, V]]
	reverse    *index[V, *entry[K, V]]
	head, tail *entry[K, V]
	size       int
	version    uint64
}

// Init is one-shot: changing an index's comparator after insertion is invalid.
func (m *Core[K, V, M]) Init(keyOrder compare.Comparator[K], valueOrder compare.Comparator[V]) {
	if m.forward != nil {
		panic("maps: already initialized")
	}
	var kind M
	tree, bidi := kind.flags()
	if tree && (keyOrder == nil || (bidi && valueOrder == nil)) {
		panic("maps: tree constructors require non-nil comparators")
	}
	if !tree && (keyOrder != nil || valueOrder != nil) {
		panic("maps: hash maps do not use comparators")
	}
	m.forward = newIndex[K, *entry[K, V]](keyOrder)
	if bidi {
		m.reverse = newIndex[V, *entry[K, V]](valueOrder)
	}
}
func (m *Core[K, V, M]) ensure() {
	if m.forward == nil {
		m.Init(nil, nil)
	}
}
func (m *Core[K, V, M]) Size() int   { return m.size }
func (m *Core[K, V, M]) Empty() bool { return m.size == 0 }
func (m *Core[K, V, M]) Put(key K, value V) {
	m.ensure()
	m.forward.validate(key)
	if m.reverse != nil {
		m.reverse.validate(value)
	}
	old, _ := m.forward.get(key)
	var owner *entry[K, V]
	if m.reverse != nil {
		owner, _ = m.reverse.get(value)
	}
	// Resolve both conflicts before altering either index. User callbacks must
	// be pure, deterministic and non-panicking throughout a mutation.
	if owner != nil && owner != old {
		m.unlink(owner)
	}
	if old != nil {
		if m.reverse != nil {
			m.reverse.remove(old.value)
		}
		old.value = value
		if m.reverse != nil {
			m.reverse.put(value, old)
		}
	} else {
		n := &entry[K, V]{key: key, value: value, previous: m.tail}
		m.forward.put(key, n)
		if m.reverse != nil {
			m.reverse.put(value, n)
		}
		if m.tail == nil {
			m.head = n
		} else {
			m.tail.next = n
		}
		m.tail = n
		m.size++
	}
	m.version++
}
func (m *Core[K, V, M]) Get(key K) (V, bool) {
	m.ensure()
	m.forward.validate(key)
	if n, ok := m.forward.get(key); ok {
		return n.value, true
	}
	var zero V
	return zero, false
}
func (m *Core[K, V, M]) ContainsKey(key K) bool { _, ok := m.Get(key); return ok }
func (m *Core[K, V, M]) LookupValue(value V) (K, bool) {
	m.ensure()
	if m.reverse == nil {
		panic("maps: reverse lookup requires a bidirectional map")
	}
	m.reverse.validate(value)
	if n, ok := m.reverse.get(value); ok {
		return n.key, true
	}
	var zero K
	return zero, false
}
func (m *Core[K, V, M]) ContainsValue(value V) bool {
	m.ensure()
	if m.reverse != nil {
		_, ok := m.LookupValue(value)
		return ok
	}
	validate(value)
	for _, current := range m.Seq() {
		validate(current)
		if base.Equals(current, value) {
			return true
		}
	}
	return false
}
func (m *Core[K, V, M]) unlink(n *entry[K, V]) {
	m.forward.remove(n.key)
	if m.reverse != nil {
		m.reverse.remove(n.value)
	}
	if n.previous == nil {
		m.head = n.next
	} else {
		n.previous.next = n.next
	}
	if n.next == nil {
		m.tail = n.previous
	} else {
		n.next.previous = n.previous
	}
	m.size--
}
func (m *Core[K, V, M]) Remove(key K) bool {
	m.ensure()
	m.forward.validate(key)
	n, ok := m.forward.get(key)
	if !ok {
		return false
	}
	m.unlink(n)
	m.version++
	return true
}
func (m *Core[K, V, M]) Clear() {
	m.ensure()
	if m.Empty() {
		return
	}
	m.forward = newIndex[K, *entry[K, V]](m.forward.order)
	if m.reverse != nil {
		m.reverse = newIndex[V, *entry[K, V]](m.reverse.order)
	}
	m.head, m.tail, m.size = nil, nil, 0
	m.version++
}
func (m *Core[K, V, M]) check(version uint64) {
	if version != m.version {
		panic("maps: iterator invalidated by mutation")
	}
}

// Seq captures a version, not a deep snapshot. Every successful Put invalidates
// it, including value replacement; callbacks are checked even on early stop.
func (m *Core[K, V, M]) Seq() iter.Seq2[K, V] {
	m.ensure()
	version := m.version
	return func(yield func(K, V) bool) {
		m.check(version)
		if m.forward.tree != nil {
			it := m.forward.tree.Iterator()
			for it.Next() {
				n := it.Value()
				more := yield(n.key, n.value)
				m.check(version)
				if !more {
					return
				}
			}
			return
		}
		for n := m.head; n != nil; n = n.next {
			more := yield(n.key, n.value)
			m.check(version)
			if !more {
				return
			}
		}
	}
}
func (m *Core[K, V, M]) Entries() []maps.Entry[K, V] {
	result := make([]maps.Entry[K, V], 0, m.size)
	for key, value := range m.Seq() {
		result = append(result, maps.Entry[K, V]{Key: key, Value: value})
	}
	return result
}
func (m *Core[K, V, M]) Keys() []K {
	result := make([]K, 0, m.size)
	for key := range m.Seq() {
		result = append(result, key)
	}
	return result
}
func (m *Core[K, V, M]) Values() []V {
	result := make([]V, 0, m.size)
	for _, value := range m.Seq() {
		result = append(result, value)
	}
	return result
}
func (m *Core[K, V, M]) Each(f func(K, V))          { containers.Each(m.Seq(), f) }
func (m *Core[K, V, M]) Any(f func(K, V) bool) bool { return containers.Any(m.Seq(), f) }
func (m *Core[K, V, M]) All(f func(K, V) bool) bool { return containers.All(m.Seq(), f) }
func (m *Core[K, V, M]) Find(f func(K, V) bool) (K, V, bool) {
	return containers.Find(m.Seq(), f)
}
func (m *Core[K, V, M]) String() string               { return fmt.Sprint(m.Entries()) }
func (m *Core[K, V, M]) ToString() string             { return m.String() }
func (m *Core[K, V, M]) ToJSON() ([]byte, error)      { return m.MarshalJSON() }
func (m *Core[K, V, M]) FromJSON(data []byte) error   { return m.UnmarshalJSON(data) }
func (m *Core[K, V, M]) MarshalJSON() ([]byte, error) { return json.Marshal(m.Entries()) }
func (m *Core[K, V, M]) Fresh() Core[K, V, M] {
	m.ensure()
	var result Core[K, V, M]
	var valueOrder compare.Comparator[V]
	if m.reverse != nil {
		valueOrder = m.reverse.order
	}
	result.Init(m.forward.order, valueOrder)
	return result
}
func (m *Core[K, V, M]) UnmarshalJSON(data []byte) (err error) {
	defer func() {
		if cause := recover(); cause != nil {
			if contract, ok := cause.(ContractError); ok {
				err = contract
			} else {
				panic(cause)
			}
		}
	}()
	var values []maps.Entry[K, V]
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	temporary := m.Fresh()
	for _, value := range values {
		temporary.Put(value.Key, value.Value)
	}
	temporary.version = m.version + 1
	*m = temporary
	return nil
}
