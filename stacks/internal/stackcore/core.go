package stackcore

import (
	"encoding/json"
	"fmt"
	"iter"

	"github.com/Shenghui-Waa/go.repo/containers"
)

type Kind interface{ linked() bool }
type Array struct{}
type Linked struct{}

func (Array) linked() bool  { return false }
func (Linked) linked() bool { return true }

// Core holds storage and mutation state together. Its zero value is usable.
// Do not copy after first use; snapshots share only referenced element objects.
type Core[T any, K Kind] struct {
	data    storage[T]
	version uint64
}

func (s *Core[T, K]) ready() {
	if s.data != nil {
		return
	}
	var kind K
	if kind.linked() {
		s.data = &linked[T]{}
	} else {
		s.data = &array[T]{}
	}
}
func (s *Core[T, K]) Push(value T) { s.ready(); s.data.push(value); s.version++ }
func (s *Core[T, K]) Pop() (T, bool) {
	s.ready()
	value, ok := s.data.pop()
	if ok {
		s.version++
	}
	return value, ok
}
func (s *Core[T, K]) Peek() (T, bool) { s.ready(); return s.data.peek() }
func (s *Core[T, K]) Size() int       { s.ready(); return s.data.len() }
func (s *Core[T, K]) Empty() bool     { return s.Size() == 0 }
func (s *Core[T, K]) Clear() {
	if !s.Empty() {
		s.data.clear()
		s.version++
	}
}
func (s *Core[T, K]) Values() []T      { s.ready(); return s.data.values() }
func (s *Core[T, K]) String() string   { return fmt.Sprint(s.Values()) }
func (s *Core[T, K]) ToString() string { return s.String() }
func (s *Core[T, K]) check(version uint64) {
	if s.version != version {
		panic("stacks: traversal invalidated by mutation")
	}
}

// Seq captures a top-first shallow snapshot and version at creation.
func (s *Core[T, K]) Seq() iter.Seq2[int, T] {
	values := s.Values()
	version := s.version
	return func(yield func(int, T) bool) {
		s.check(version)
		for i, value := range values {
			more := yield(i, value)
			s.check(version)
			if !more {
				return
			}
		}
	}
}
func (s *Core[T, K]) Each(f func(int, T))          { containers.Each(s.Seq(), f) }
func (s *Core[T, K]) Any(f func(int, T) bool) bool { return containers.Any(s.Seq(), f) }
func (s *Core[T, K]) All(f func(int, T) bool) bool { return containers.All(s.Seq(), f) }
func (s *Core[T, K]) Find(f func(int, T) bool) (int, T, bool) {
	i, value, ok := containers.Find(s.Seq(), f)
	if !ok {
		i = -1
	}
	return i, value, ok
}
func fromValues[T any, K Kind](values []T) Core[T, K] {
	var result Core[T, K]
	// Input is top-first, so build from bottom to top.
	for i := len(values) - 1; i >= 0; i-- {
		result.Push(values[i])
	}
	return result
}
func (s *Core[T, K]) Map(f func(int, T) T) Core[T, K] {
	if f == nil {
		panic("stacks: nil transform")
	}
	values := make([]T, 0, s.Size())
	s.Each(func(i int, value T) { values = append(values, f(i, value)) })
	return fromValues[T, K](values)
}
func (s *Core[T, K]) Select(f func(int, T) bool) Core[T, K] {
	if f == nil {
		panic("stacks: nil predicate")
	}
	values := make([]T, 0, s.Size())
	s.Each(func(i int, value T) {
		if f(i, value) {
			values = append(values, value)
		}
	})
	return fromValues[T, K](values)
}
func (s *Core[T, K]) MarshalJSON() ([]byte, error) { return json.Marshal(s.Values()) }
func (s *Core[T, K]) ToJSON() ([]byte, error)      { return s.MarshalJSON() }
func (s *Core[T, K]) FromJSON(data []byte) error   { return s.UnmarshalJSON(data) }
func (s *Core[T, K]) UnmarshalJSON(data []byte) error {
	var values []T
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	result := fromValues[T, K](values)
	result.version = s.version + 1
	*s = result
	return nil
}
