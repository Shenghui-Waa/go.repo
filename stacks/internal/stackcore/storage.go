package stackcore

type storage[T any] interface {
	push(T)
	pop() (T, bool)
	peek() (T, bool)
	len() int
	clear()
	values() []T
}

type array[T any] struct{ items []T }

func (a *array[T]) len() int     { return len(a.items) }
func (a *array[T]) push(value T) { a.items = append(a.items, value) }
func (a *array[T]) pop() (T, bool) {
	var zero T
	if len(a.items) == 0 {
		return zero, false
	}
	last := len(a.items) - 1
	value := a.items[last]
	a.items[last] = zero
	a.items = a.items[:last]
	return value, true
}
func (a *array[T]) peek() (T, bool) {
	if len(a.items) == 0 {
		var zero T
		return zero, false
	}
	return a.items[len(a.items)-1], true
}
func (a *array[T]) clear() { a.items = nil }
func (a *array[T]) values() []T {
	result := make([]T, len(a.items))
	for i := range result {
		result[i] = a.items[len(a.items)-1-i]
	}
	return result
}

type node[T any] struct {
	value T
	next  *node[T]
}
type linked[T any] struct {
	top  *node[T]
	size int
}

func (l *linked[T]) len() int     { return l.size }
func (l *linked[T]) push(value T) { l.top = &node[T]{value: value, next: l.top}; l.size++ }
func (l *linked[T]) pop() (T, bool) {
	if l.top == nil {
		var zero T
		return zero, false
	}
	value := l.top.value
	l.top = l.top.next
	l.size--
	return value, true
}
func (l *linked[T]) peek() (T, bool) {
	if l.top == nil {
		var zero T
		return zero, false
	}
	return l.top.value, true
}
func (l *linked[T]) clear() { l.top, l.size = nil, 0 }
func (l *linked[T]) values() []T {
	result := make([]T, 0, l.size)
	for n := l.top; n != nil; n = n.next {
		result = append(result, n.value)
	}
	return result
}
