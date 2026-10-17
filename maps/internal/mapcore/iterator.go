package mapcore

import "github.com/Shenghui-Waa/go.repo/maps"

// Cursor captures shallow entries in O(n) space, then traverses in O(1) per step.
// It still fails fast on all source mutations, including overwritten values.
type Cursor[K, V any] struct {
	entries []maps.Entry[K, V]
	index   int
	check   func()
}

func (m *Core[K, V, M]) Iterator() *Cursor[K, V] {
	version := m.version
	return &Cursor[K, V]{
		entries: m.Entries(), index: -1,
		check: func() { m.check(version) },
	}
}
func (c *Cursor[K, V]) Next() bool {
	c.check()
	if c.index < len(c.entries) {
		c.index++
	}
	return c.index < len(c.entries)
}
func (c *Cursor[K, V]) Prev() bool {
	c.check()
	if c.index >= 0 {
		c.index--
	}
	return c.index >= 0
}
func (c *Cursor[K, V]) current() maps.Entry[K, V] {
	c.check()
	if c.index < 0 || c.index >= len(c.entries) {
		panic("maps: invalid cursor position")
	}
	return c.entries[c.index]
}
func (c *Cursor[K, V]) Key() K      { return c.current().Key }
func (c *Cursor[K, V]) Value() V    { return c.current().Value }
func (c *Cursor[K, V]) Begin()      { c.check(); c.index = -1 }
func (c *Cursor[K, V]) End()        { c.check(); c.index = len(c.entries) }
func (c *Cursor[K, V]) First() bool { c.Begin(); return c.Next() }
func (c *Cursor[K, V]) Last() bool  { c.End(); return c.Prev() }
func (c *Cursor[K, V]) NextTo(f func(K, V) bool) bool {
	if f == nil {
		panic("maps: nil predicate")
	}
	for c.Next() {
		found := f(c.Key(), c.Value())
		c.check()
		if found {
			return true
		}
	}
	return false
}
func (c *Cursor[K, V]) PrevTo(f func(K, V) bool) bool {
	if f == nil {
		panic("maps: nil predicate")
	}
	for c.Prev() {
		found := f(c.Key(), c.Value())
		c.check()
		if found {
			return true
		}
	}
	return false
}

// Bounds is available on tree wrappers. Hash wrappers deliberately omit it.
func (m *Core[K, V, M]) Bound(key K, mode string) (K, V, bool) {
	m.ensure()
	tree := m.forward.tree
	if tree == nil {
		panic("maps: ordered bounds require a tree")
	}
	var n *entry[K, V]
	var found bool
	switch mode {
	case "min":
		_, n, found = tree.Min()
	case "max":
		_, n, found = tree.Max()
	case "floor":
		m.forward.validate(key)
		_, n, found = tree.Floor(key)
	case "ceiling":
		m.forward.validate(key)
		_, n, found = tree.Ceiling(key)
	default:
		panic("maps: unknown bound mode")
	}
	if !found {
		var k K
		var v V
		return k, v, false
	}
	return n.key, n.value, true
}

func (m *Core[K, V, M]) Map(f func(K, V) (K, V)) Core[K, V, M] {
	if f == nil {
		panic("maps: nil transform")
	}
	result := m.Fresh()
	m.Each(func(key K, value V) { k, v := f(key, value); result.Put(k, v) })
	return result
}
func (m *Core[K, V, M]) Select(f func(K, V) bool) Core[K, V, M] {
	if f == nil {
		panic("maps: nil predicate")
	}
	result := m.Fresh()
	m.Each(func(key K, value V) {
		if f(key, value) {
			result.Put(key, value)
		}
	})
	return result
}
