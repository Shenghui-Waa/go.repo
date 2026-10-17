package ordered

import (
	"encoding/json"
	"fmt"
	"iter"

	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/trees"
	avl "github.com/Shenghui-Waa/go.repo/trees/internal/avltree"
	bt "github.com/Shenghui-Waa/go.repo/trees/internal/btree"
	rb "github.com/Shenghui-Waa/go.repo/trees/internal/redblacktree"
)

type Kind interface{ kind() int }
type AVL struct{}
type RedBlack struct{}
type BTree struct{}

func (AVL) kind() int      { return 0 }
func (RedBlack) kind() int { return 1 }
func (BTree) kind() int    { return 2 }

type backend[K, V any] interface {
	Put(K, V)
	Get(K) (V, bool)
	Remove(K)
	Keys() []K
	Values() []V
	Size() int
	Clear()
}

// Core keeps nodes and comparator private. Construct once; do not copy after use.
type Core[K, V any, M Kind] struct {
	data    backend[K, V]
	order   compare.Comparator[K]
	degree  int
	version uint64
}

func New[K, V any, M Kind](order compare.Comparator[K], degree int) Core[K, V, M] {
	if order == nil {
		panic("trees: nil comparator")
	}
	result := Core[K, V, M]{order: order, degree: degree}
	var kind M
	switch kind.kind() {
	case 0:
		result.data = avl.NewWith[K, V](order)
	case 1:
		result.data = rb.NewWith[K, V](order)
	case 2:
		result.data = bt.NewWith[K, V](degree, order)
	}
	return result
}
func (t *Core[K, V, M]) ready() {
	if t.data == nil {
		panic("trees: use a constructor before operating on a tree")
	}
}
func (t *Core[K, V, M]) validate(key K) {
	t.ready()
	if t.order(key, key) != 0 {
		panic("trees: comparator is not reflexive")
	}
}
func (t *Core[K, V, M]) Size() int   { t.ready(); return t.data.Size() }
func (t *Core[K, V, M]) Empty() bool { return t.Size() == 0 }
func (t *Core[K, V, M]) Order() int  { t.ready(); return t.degree }
func (t *Core[K, V, M]) Put(key K, value V) {
	t.validate(key)
	// Keep the original representative for comparator-equivalent keys.
	if original, _, ok := t.bound(key, true); ok && t.order(key, original) == 0 {
		key = original
	}
	t.data.Put(key, value)
	t.version++
}
func (t *Core[K, V, M]) Get(key K) (V, bool)    { t.validate(key); return t.data.Get(key) }
func (t *Core[K, V, M]) ContainsKey(key K) bool { _, ok := t.Get(key); return ok }
func (t *Core[K, V, M]) Remove(key K) bool {
	t.validate(key)
	if _, ok := t.data.Get(key); !ok {
		return false
	}
	t.data.Remove(key)
	t.version++
	return true
}
func (t *Core[K, V, M]) Clear() {
	if !t.Empty() {
		t.data.Clear()
		t.version++
	}
}
func (t *Core[K, V, M]) Keys() []K   { t.ready(); return t.data.Keys() }
func (t *Core[K, V, M]) Values() []V { t.ready(); return t.data.Values() }

func (t *Core[K, V, M]) rawIterator() containers.ReverseIteratorWithKey[K, V] {
	t.ready()
	switch data := t.data.(type) {
	case *avl.Tree[K, V]:
		return data.Iterator()
	case *rb.Tree[K, V]:
		return data.Iterator()
	case *bt.Tree[K, V]:
		return data.Iterator()
	default:
		panic("trees: unknown backend")
	}
}
func (t *Core[K, V, M]) check(version uint64) {
	if t.version != version {
		panic("trees: traversal invalidated by mutation")
	}
}
func (t *Core[K, V, M]) Seq() iter.Seq2[K, V] {
	t.ready()
	version := t.version
	return func(yield func(K, V) bool) {
		t.check(version)
		cursor := t.rawIterator()
		for cursor.Next() {
			more := yield(cursor.Key(), cursor.Value())
			t.check(version)
			if !more {
				return
			}
		}
	}
}
func (t *Core[K, V, M]) Entries() []trees.Entry[K, V] {
	result := make([]trees.Entry[K, V], 0, t.Size())
	for key, value := range t.Seq() {
		result = append(result, trees.Entry[K, V]{Key: key, Value: value})
	}
	return result
}
func (t *Core[K, V, M]) Each(f func(K, V))          { containers.Each(t.Seq(), f) }
func (t *Core[K, V, M]) Any(f func(K, V) bool) bool { return containers.Any(t.Seq(), f) }
func (t *Core[K, V, M]) All(f func(K, V) bool) bool { return containers.All(t.Seq(), f) }
func (t *Core[K, V, M]) Find(f func(K, V) bool) (K, V, bool) {
	return containers.Find(t.Seq(), f)
}
func (t *Core[K, V, M]) Min() (K, V, bool) {
	i := t.rawIterator()
	if i.First() {
		return i.Key(), i.Value(), true
	}
	var key K
	var value V
	return key, value, false
}
func (t *Core[K, V, M]) Max() (K, V, bool) {
	i := t.rawIterator()
	if i.Last() {
		return i.Key(), i.Value(), true
	}
	var key K
	var value V
	return key, value, false
}
func (t *Core[K, V, M]) Floor(key K) (K, V, bool)   { t.validate(key); return t.bound(key, true) }
func (t *Core[K, V, M]) Ceiling(key K) (K, V, bool) { t.validate(key); return t.bound(key, false) }
func (t *Core[K, V, M]) bound(key K, floor bool) (K, V, bool) {
	var k K
	var v V
	found := false
	accept := func(nk K, nv V) bool {
		c := t.order(key, nk)
		if c == 0 || (floor && c > 0) || (!floor && c < 0) {
			k, v, found = nk, nv, true
		}
		return c < 0
	}
	switch data := t.data.(type) {
	case *avl.Tree[K, V]:
		for n := data.Root; n != nil; {
			left := accept(n.Key, n.Value)
			if t.order(key, n.Key) == 0 {
				break
			}
			if left {
				n = n.Children[0]
			} else {
				n = n.Children[1]
			}
		}
	case *rb.Tree[K, V]:
		for n := data.Root; n != nil; {
			left := accept(n.Key, n.Value)
			if t.order(key, n.Key) == 0 {
				break
			}
			if left {
				n = n.Left
			} else {
				n = n.Right
			}
		}
	case *bt.Tree[K, V]:
		for n := data.Root; n != nil; {
			low, high := 0, len(n.Entries)
			for low < high {
				mid := low + (high-low)/2
				if t.order(n.Entries[mid].Key, key) < 0 {
					low = mid + 1
				} else {
					high = mid
				}
			}
			if low < len(n.Entries) && t.order(n.Entries[low].Key, key) == 0 {
				e := n.Entries[low]
				return e.Key, e.Value, true
			}
			candidate := low
			if floor {
				candidate--
			}
			if candidate >= 0 && candidate < len(n.Entries) {
				e := n.Entries[candidate]
				k, v, found = e.Key, e.Value, true
			}
			if len(n.Children) == 0 {
				break
			}
			n = n.Children[low]
		}
	}
	return k, v, found
}
func (t *Core[K, V, M]) String() string               { return fmt.Sprint(t.Entries()) }
func (t *Core[K, V, M]) ToString() string             { return t.String() }
func (t *Core[K, V, M]) MarshalJSON() ([]byte, error) { return json.Marshal(t.Entries()) }
func (t *Core[K, V, M]) ToJSON() ([]byte, error)      { return t.MarshalJSON() }
func (t *Core[K, V, M]) FromJSON(data []byte) error   { return t.UnmarshalJSON(data) }
func (t *Core[K, V, M]) UnmarshalJSON(data []byte) error {
	t.ready()
	var entries []trees.Entry[K, V]
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	result := New[K, V, M](t.order, t.degree)
	for _, e := range entries {
		result.Put(e.Key, e.Value)
	}
	result.version = t.version + 1
	*t = result
	return nil
}
func (t *Core[K, V, M]) Map(f func(K, V) (K, V)) Core[K, V, M] {
	if f == nil {
		panic("trees: nil transform")
	}
	t.ready()
	result := New[K, V, M](t.order, t.degree)
	t.Each(func(k K, v V) { key, value := f(k, v); result.Put(key, value) })
	return result
}
func (t *Core[K, V, M]) Select(f func(K, V) bool) Core[K, V, M] {
	if f == nil {
		panic("trees: nil predicate")
	}
	t.ready()
	result := New[K, V, M](t.order, t.degree)
	t.Each(func(k K, v V) {
		if f(k, v) {
			result.Put(k, v)
		}
	})
	return result
}
