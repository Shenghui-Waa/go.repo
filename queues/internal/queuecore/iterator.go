package queuecore

// Iterator is an O(n) snapshot cursor with O(1) steps and fail-fast checks.
type Iterator[T any] struct {
	values []T
	index  int
	check  func()
}

func (q *Core[T, K]) Iterator() *Iterator[T] {
	values := q.Values()
	version := q.version
	return &Iterator[T]{values: values, index: -1, check: func() { q.check(version) }}
}
func (i *Iterator[T]) Next() bool {
	i.check()
	if i.index < len(i.values) {
		i.index++
	}
	return i.index < len(i.values)
}
func (i *Iterator[T]) Prev() bool {
	i.check()
	if i.index >= 0 {
		i.index--
	}
	return i.index >= 0
}
func (i *Iterator[T]) Index() int {
	i.check()
	if i.index < 0 || i.index >= len(i.values) {
		panic("queues: invalid cursor position")
	}
	return i.index
}
func (i *Iterator[T]) Value() T    { return i.values[i.Index()] }
func (i *Iterator[T]) Begin()      { i.check(); i.index = -1 }
func (i *Iterator[T]) End()        { i.check(); i.index = len(i.values) }
func (i *Iterator[T]) First() bool { i.Begin(); return i.Next() }
func (i *Iterator[T]) Last() bool  { i.End(); return i.Prev() }
func (i *Iterator[T]) NextTo(f func(int, T) bool) bool {
	if f == nil {
		panic("queues: nil predicate")
	}
	for i.Next() {
		found := f(i.Index(), i.Value())
		i.check()
		if found {
			return true
		}
	}
	return false
}
func (i *Iterator[T]) PrevTo(f func(int, T) bool) bool {
	if f == nil {
		panic("queues: nil predicate")
	}
	for i.Prev() {
		found := f(i.Index(), i.Value())
		i.check()
		if found {
			return true
		}
	}
	return false
}
