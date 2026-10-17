package ordered

import "github.com/Shenghui-Waa/go.repo/containers"

type Iterator[K, V any] struct {
	cursor containers.ReverseIteratorWithKey[K, V]
	check  func()
	valid  bool
}

func (t *Core[K, V, M]) Iterator() *Iterator[K, V] {
	version := t.version
	return &Iterator[K, V]{cursor: t.rawIterator(), check: func() { t.check(version) }}
}
func (i *Iterator[K, V]) Next() bool { i.check(); i.valid = i.cursor.Next(); return i.valid }
func (i *Iterator[K, V]) Prev() bool { i.check(); i.valid = i.cursor.Prev(); return i.valid }
func (i *Iterator[K, V]) position() {
	i.check()
	if !i.valid {
		panic("trees: invalid iterator position")
	}
}
func (i *Iterator[K, V]) Key() K      { i.position(); return i.cursor.Key() }
func (i *Iterator[K, V]) Value() V    { i.position(); return i.cursor.Value() }
func (i *Iterator[K, V]) Begin()      { i.check(); i.cursor.Begin(); i.valid = false }
func (i *Iterator[K, V]) End()        { i.check(); i.cursor.End(); i.valid = false }
func (i *Iterator[K, V]) First() bool { i.Begin(); return i.Next() }
func (i *Iterator[K, V]) Last() bool  { i.End(); return i.Prev() }
func (i *Iterator[K, V]) NextTo(f func(K, V) bool) bool {
	if f == nil {
		panic("trees: nil predicate")
	}
	for i.Next() {
		found := f(i.Key(), i.Value())
		i.check()
		if found {
			return true
		}
	}
	return false
}
func (i *Iterator[K, V]) PrevTo(f func(K, V) bool) bool {
	if f == nil {
		panic("trees: nil predicate")
	}
	for i.Prev() {
		found := f(i.Key(), i.Value())
		i.check()
		if found {
			return true
		}
	}
	return false
}
