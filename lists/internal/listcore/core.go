package listcore

import (
	"encoding/json"
	"fmt"
	"iter"
	"slices"

	"github.com/Shenghui-Waa/go.repo/base"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/lists"
)

// Core holds shared behavior; all mutations keep storage and version together.
// The zero value is an empty array-backed list. Public wrappers select storage
// through a generic marker, including for their own zero values.
type Core[T any, K Kind] struct {
	data    storage[T]
	order   compare.Comparator[T]
	version uint64
}

type Kind interface{ kind() int }
type Array struct{}
type Single struct{}
type Double struct{}

func (Array) kind() int  { return 0 }
func (Single) kind() int { return 1 }
func (Double) kind() int { return 2 }

func (l *Core[T, K]) store() storage[T] {
	if l.data == nil {
		var kind K
		switch kind.kind() {
		case 1:
			l.data = &single[T]{}
		case 2:
			l.data = &double[T]{}
		default:
			l.data = &array[T]{}
		}
	}
	return l.data
}

func (l *Core[T, K]) Configure(order compare.Comparator[T]) {
	if order == nil {
		panic("lists: nil comparator")
	}
	l.order = order
}
func (l *Core[T, K]) Comparator() compare.Comparator[T] { return l.order }
func (l *Core[T, K]) Size() int                         { return l.store().len() }
func (l *Core[T, K]) Empty() bool                       { return l.Size() == 0 }
func (l *Core[T, K]) Get(i int) (T, bool) {
	if i < 0 || i >= l.Size() {
		var zero T
		return zero, false
	}
	return l.store().get(i), true
}
func (l *Core[T, K]) Add(values ...T)     { l.Insert(l.Size(), values...) }
func (l *Core[T, K]) Append(values ...T)  { l.Add(values...) }
func (l *Core[T, K]) Prepend(values ...T) { l.Insert(0, values...) }
func (l *Core[T, K]) Insert(i int, values ...T) bool {
	if i < 0 || i > l.Size() {
		return false
	}
	if len(values) != 0 {
		l.store().insert(i, values)
		l.version++
	}
	return true
}
func (l *Core[T, K]) Set(i int, value T) bool {
	if i < 0 || i >= l.Size() {
		return false
	}
	l.store().set(i, value)
	l.version++
	return true
}
func (l *Core[T, K]) Remove(i int) bool {
	if i < 0 || i >= l.Size() {
		return false
	}
	l.store().remove(i)
	l.version++
	return true
}
func (l *Core[T, K]) RemoveValue(value T) bool { return l.Remove(l.IndexOf(value)) }
func (l *Core[T, K]) Clear() {
	if !l.Empty() {
		l.store().clear()
		l.version++
	}
}
func (l *Core[T, K]) Values() []T {
	values := make([]T, 0, l.Size())
	for _, value := range l.Seq() {
		values = append(values, value)
	}
	return values
}
func (l *Core[T, K]) Contains(values ...T) bool {
	requireValues(values)
	for _, value := range values {
		if l.IndexOf(value) < 0 {
			return false
		}
	}
	return true
}
func (l *Core[T, K]) IndexOf(value T) int {
	requireValues([]T{value})
	for i, current := range l.Seq() {
		requireValues([]T{current})
		if base.Equals(current, value) {
			return i
		}
	}
	return -1
}
func (l *Core[T, K]) LastIndexOf(value T) int {
	requireValues([]T{value})
	index := -1
	for i, current := range l.Seq() {
		requireValues([]T{current})
		if base.Equals(current, value) {
			index = i
		}
	}
	return index
}
func (l *Core[T, K]) Swap(i, j int) bool {
	if i < 0 || j < 0 || i >= l.Size() || j >= l.Size() {
		return false
	}
	if i != j {
		a, b := l.store().get(i), l.store().get(j)
		l.store().set(i, b)
		l.store().set(j, a)
		l.version++
	}
	return true
}
func (l *Core[T, K]) Sort(orders ...compare.Comparator[T]) {
	if len(orders) > 1 {
		panic("lists: expected at most one comparator")
	}
	order := l.order
	if len(orders) == 1 {
		order = orders[0]
	}
	if order == nil {
		panic("lists: Sort requires a comparator")
	}
	version := l.version
	values := l.Values()
	slices.SortStableFunc(values, order)
	l.check(version)
	if len(values) > 1 {
		l.replace(values)
	}
}
func (l *Core[T, K]) replace(values []T) {
	l.store().clear()
	l.store().insert(0, values)
	l.version++
}
func (l *Core[T, K]) check(version uint64) {
	if version != l.version {
		panic("lists: iterator invalidated by mutation")
	}
}

// Seq captures a version at creation and checks it even after a final callback.
// Set, Swap and Sort invalidate traversal as well as structural changes.
func (l *Core[T, K]) Seq() iter.Seq2[int, T] {
	version := l.version
	return func(yield func(int, T) bool) {
		l.check(version)
		read := l.store().reader()
		for i := 0; i < l.Size(); i++ {
			more := yield(i, read(i))
			l.check(version)
			if !more {
				return
			}
		}
	}
}
func (l *Core[T, K]) Each(f func(int, T))          { containers.Each(l.Seq(), f) }
func (l *Core[T, K]) Any(f func(int, T) bool) bool { return containers.Any(l.Seq(), f) }
func (l *Core[T, K]) All(f func(int, T) bool) bool { return containers.All(l.Seq(), f) }
func (l *Core[T, K]) Find(f func(int, T) bool) (int, T, bool) {
	i, value, ok := containers.Find(l.Seq(), f)
	if !ok {
		i = -1
	}
	return i, value, ok
}
func (l *Core[T, K]) Equal(other any) bool {
	if base.IsNil(other) {
		return false
	}
	right, ok := other.(interface {
		lists.List[T]
		base.Equaler
		base.Hasher
	})
	if !ok || l.Size() != right.Size() {
		return false
	}
	values := right.Values()
	for i, value := range l.Seq() {
		requireValues([]T{value, values[i]})
		if !base.Equals(value, values[i]) {
			return false
		}
	}
	return true
}
func (l *Core[T, K]) HashCode() uint64 {
	result := uint64(1)
	for _, value := range l.Seq() {
		requireValues([]T{value})
		result = 31*result + base.HashCode(value)
	}
	return result
}
func (l *Core[T, K]) String() string               { return fmt.Sprint(l.Values()) }
func (l *Core[T, K]) ToString() string             { return l.String() }
func (l *Core[T, K]) ToJSON() ([]byte, error)      { return l.MarshalJSON() }
func (l *Core[T, K]) FromJSON(data []byte) error   { return l.UnmarshalJSON(data) }
func (l *Core[T, K]) MarshalJSON() ([]byte, error) { return json.Marshal(l.Values()) }
func (l *Core[T, K]) UnmarshalJSON(data []byte) error {
	var values []T
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	l.replace(values)
	return nil
}
