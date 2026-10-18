package singlylinkedlist

// AddFirst inserts one value at the head in O(1).
func (l *List[T]) AddFirst(value T) { l.Prepend(value) }

// AddLast appends one value at the tail in O(1).
func (l *List[T]) AddLast(value T) { l.Append(value) }

// GetFirst reads the head in O(1), returning false when empty.
func (l *List[T]) GetFirst() (T, bool) { return l.Get(0) }

// GetLast reads the tail in O(1), returning false when empty.
func (l *List[T]) GetLast() (T, bool) { return l.Get(l.Size() - 1) }

// RemoveFirst removes and returns the head in O(1).
func (l *List[T]) RemoveFirst() (T, bool) {
	value, ok := l.GetFirst()
	if ok {
		l.Remove(0)
	}
	return value, ok
}

// RemoveLast removes and returns the tail in O(n), locating its predecessor.
func (l *List[T]) RemoveLast() (T, bool) {
	value, ok := l.GetLast()
	if ok {
		l.Remove(l.Size() - 1)
	}
	return value, ok
}
