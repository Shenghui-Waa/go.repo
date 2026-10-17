package binaryheap

// Iterator walks heap layout, not priority order.
type Iterator[T any] struct {
	heap    *Heap[T]
	version uint64
	index   int
}

func (h *Heap[T]) Iterator() *Iterator[T] {
	h.ready()
	return &Iterator[T]{heap: h, version: h.version, index: -1}
}
func (i *Iterator[T]) Next() bool {
	i.heap.check(i.version)
	if i.index < i.heap.Size() {
		i.index++
	}
	return i.index < i.heap.Size()
}
func (i *Iterator[T]) Index() int {
	i.heap.check(i.version)
	if i.index < 0 || i.index >= i.heap.Size() {
		panic("binaryheap: invalid cursor")
	}
	return i.index
}
func (i *Iterator[T]) Value() T    { return i.heap.values[i.Index()] }
func (i *Iterator[T]) Begin()      { i.heap.check(i.version); i.index = -1 }
func (i *Iterator[T]) First() bool { i.Begin(); return i.Next() }
func (i *Iterator[T]) NextTo(f func(int, T) bool) bool {
	if f == nil {
		panic("binaryheap: nil predicate")
	}
	for i.Next() {
		found := f(i.Index(), i.Value())
		i.heap.check(i.version)
		if found {
			return true
		}
	}
	return false
}
