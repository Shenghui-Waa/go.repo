package setcore

import (
	"encoding/json"
	"fmt"
	"iter"

	"github.com/Shenghui-Waa/go.repo/base"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/maps"
	"github.com/Shenghui-Waa/go.repo/maps/hashmap"
	"github.com/Shenghui-Waa/go.repo/maps/linkedhashmap"
	"github.com/Shenghui-Waa/go.repo/maps/treemap"
	"github.com/Shenghui-Waa/go.repo/sets"
)

type Kind interface{ kind() int }
type Hash struct{}
type Linked struct{}
type Tree struct{}

func (Hash) kind() int   { return 0 }
func (Linked) kind() int { return 1 }
func (Tree) kind() int   { return 2 }

// Core owns the backing map. Never copy a populated Core.
type backingMap[T any] interface {
	maps.Map[T, struct{}]
	containers.Clearer
	containers.KeyedSequence[T, struct{}]
	ContainsKey(T) bool
	FromJSON([]byte) error
}

type Core[T any, K Kind] struct {
	data    backingMap[T]
	order   compare.Comparator[T]
	version uint64
}

func New[T any, K Kind](order compare.Comparator[T]) Core[T, K] {
	result := Core[T, K]{order: order}
	result.ready()
	return result
}
func (s *Core[T, K]) ready() {
	if s.data != nil {
		return
	}
	var kind K
	switch kind.kind() {
	case 0:
		s.data = hashmap.New[T, struct{}]()
	case 1:
		s.data = linkedhashmap.New[T, struct{}]()
	case 2:
		s.data = treemap.NewWithComparator[T, struct{}](s.order)
	}
}
func (s *Core[T, K]) Size() int   { s.ready(); return s.data.Size() }
func (s *Core[T, K]) Empty() bool { return s.Size() == 0 }
func (s *Core[T, K]) Add(values ...T) int {
	s.ready()
	// Validate the whole batch before mutation. User comparison/hash callbacks
	// must remain deterministic, non-panicking and must not mutate this set.
	for _, value := range values {
		s.data.ContainsKey(value)
	}
	added := 0
	for _, value := range values {
		if !s.data.ContainsKey(value) {
			s.data.Put(value, struct{}{})
			s.version++
			added++
		}
	}
	return added
}
func (s *Core[T, K]) Remove(values ...T) int {
	s.ready()
	for _, value := range values {
		s.data.ContainsKey(value)
	}
	removed := 0
	for _, value := range values {
		if s.data.Remove(value) {
			s.version++
			removed++
		}
	}
	return removed
}
func (s *Core[T, K]) Contains(values ...T) bool {
	s.ready()
	for _, value := range values {
		if !s.data.ContainsKey(value) {
			return false
		}
	}
	return true
}
func (s *Core[T, K]) Clear() {
	if !s.Empty() {
		s.data.Clear()
		s.version++
	}
}
func (s *Core[T, K]) Values() []T { s.ready(); return s.data.Keys() }
func (s *Core[T, K]) check(version uint64) {
	if s.version != version {
		panic("sets: traversal invalidated by mutation")
	}
}

// Seq returns traversal indices, not stable member identifiers.
func (s *Core[T, K]) Seq() iter.Seq2[int, T] {
	s.ready()
	version := s.version
	return func(yield func(int, T) bool) {
		s.check(version)
		index := 0
		for value := range s.data.Seq() {
			more := yield(index, value)
			s.check(version)
			if !more {
				return
			}
			index++
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
func (s *Core[T, K]) String() string    { return fmt.Sprint(s.Values()) }
func (s *Core[T, K]) ToString() string  { return s.String() }
func (s *Core[T, K]) fresh() Core[T, K] { s.ready(); return New[T, K](s.order) }
func (s *Core[T, K]) Map(f func(int, T) T) Core[T, K] {
	if f == nil {
		panic("sets: nil transform")
	}
	result := s.fresh()
	s.Each(func(i int, value T) { result.Add(f(i, value)) })
	return result
}
func (s *Core[T, K]) Select(f func(int, T) bool) Core[T, K] {
	if f == nil {
		panic("sets: nil predicate")
	}
	result := s.fresh()
	s.Each(func(i int, value T) {
		if f(i, value) {
			result.Add(value)
		}
	})
	return result
}

// Combine normalizes the other operand to receiver membership semantics.
// Both source sets remain unchanged; receiver representatives win duplicates.
func (s *Core[T, K]) Combine(other sets.Set[T], operation string) Core[T, K] {
	if base.IsNil(other) {
		panic("sets: nil operand")
	}
	right := s.fresh()
	for _, value := range other.Values() {
		right.Add(value)
	}
	result := s.fresh()
	for _, value := range s.Seq() {
		present := right.Contains(value)
		if operation == "union" || (operation == "intersection" && present) ||
			((operation == "difference" || operation == "symmetric") && !present) {
			result.Add(value)
		}
	}
	if operation == "union" || operation == "symmetric" {
		for _, value := range right.Seq() {
			if !s.Contains(value) {
				result.Add(value)
			}
		}
	}
	return result
}
func (s *Core[T, K]) IsSubsetOf(other sets.Set[T]) bool {
	result := s.Combine(other, "intersection")
	return result.Size() == s.Size()
}
func (s *Core[T, K]) IsSupersetOf(other sets.Set[T]) bool {
	if base.IsNil(other) {
		panic("sets: nil operand")
	}
	for _, value := range other.Values() {
		if !s.Contains(value) {
			return false
		}
	}
	return true
}
func (s *Core[T, K]) IsDisjoint(other sets.Set[T]) bool {
	result := s.Combine(other, "intersection")
	return result.Empty()
}
func (s *Core[T, K]) MarshalJSON() ([]byte, error) { return json.Marshal(s.Values()) }
func (s *Core[T, K]) ToJSON() ([]byte, error)      { return s.MarshalJSON() }
func (s *Core[T, K]) FromJSON(data []byte) error   { return s.UnmarshalJSON(data) }
func (s *Core[T, K]) UnmarshalJSON(data []byte) error {
	// Preserve raw elements through the map's decoder, including its contract
	// validation errors. Build and validate a temporary index before committing.
	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	entries := make([]maps.Entry[json.RawMessage, struct{}], len(values))
	for i, value := range values {
		entries[i].Key = value
	}
	encoded, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	result := s.fresh()
	if err := result.data.FromJSON(encoded); err != nil {
		return err
	}
	result.version = s.version + 1
	*s = result
	return nil
}

// Bound delegates to the tree map without exposing the backing index.
func (s *Core[T, K]) Bound(value T, mode string) (T, bool) {
	s.ready()
	tree, ok := s.data.(*treemap.Map[T, struct{}])
	if !ok {
		panic("sets: bounds require a tree set")
	}
	var result T
	var found bool
	switch mode {
	case "min":
		result, _, found = tree.Min()
	case "max":
		result, _, found = tree.Max()
	case "floor":
		result, _, found = tree.Floor(value)
	case "ceiling":
		result, _, found = tree.Ceiling(value)
	default:
		panic("sets: unknown bound")
	}
	return result, found
}
