package containers_test

import (
	"cmp"
	"fmt"
	"iter"
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
)

// Deliberately exposes storage to test helpers against aliasing sources.
type source[T any] struct{ values []T }

func (s *source[T]) Values() []T    { return s.values }
func (s *source[T]) Empty() bool    { return len(s.values) == 0 }
func (s *source[T]) Size() int      { return len(s.values) }
func (s *source[T]) Clear()         { s.values = nil }
func (s *source[T]) String() string { return fmt.Sprint(s.values) }

var _ containers.Container[int] = (*source[int])(nil)

type item struct {
	rank int
	name string
}

func (i item) CompareTo(other item) int { return cmp.Compare(i.rank, other.rank) }

func TestSorting(t *testing.T) {
	s := &source[int]{[]int{3, 1, 2}}
	got := containers.GetSortedValues[int](s)
	if !slices.Equal(got, []int{1, 2, 3}) || s.values[0] != 3 {
		t.Fatal("sorting changed the source or returned wrong order")
	}
	got[0] = 99
	if s.values[1] != 1 {
		t.Fatal("sorted result aliases the source")
	}
	for _, values := range [][]int{nil, {}, {1}} {
		small := &source[int]{values}
		result := containers.GetSortedValues[int](small)
		if (result == nil) != (values == nil) || !slices.Equal(result, values) {
			t.Fatal("empty or singleton semantics changed")
		}
		if len(result) == 1 {
			result[0] = 10
			if values[0] != 1 {
				t.Fatal("singleton result aliases source")
			}
		}
	}
	tasks := &source[item]{[]item{{2, "a"}, {1, "b"}, {1, "c"}}}
	result := containers.GetSortedComparableValues[item](tasks)
	if !slices.Equal(result, []item{{1, "b"}, {1, "c"}, {2, "a"}}) {
		t.Fatal("natural ordering must be stable")
	}
	byRank := compare.FromComparable[item]().Reversed()
	if containers.GetSortedValuesFunc[item](tasks, byRank)[0].rank != 2 {
		t.Fatal("custom ordering ignored")
	}
}

func TestSnapshot(t *testing.T) {
	s := &source[int]{[]int{1, 2}}
	seq := containers.Snapshot[int](s)
	s.values[0] = 9
	s.Clear()
	for range 2 {
		got := slices.Collect(containers.Map(seq, func(k, v int) int { return k + v }))
		if !slices.Equal(got, []int{1, 3}) {
			t.Fatal(got)
		}
	}
}

func TestEnumerable(t *testing.T) {
	seq := slices.All([]int{0, 2, 4})
	key, value, ok := containers.Find(seq, func(k, v int) bool { return v == 0 })
	if !ok || key != 0 || value != 0 {
		t.Fatal("zero match lost")
	}
	_, _, ok = containers.Find(seq, func(k, v int) bool { return false })
	if ok {
		t.Fatal("unexpected match")
	}
	calls := 0
	if !containers.Any(seq, func(k, v int) bool { calls++; return true }) || calls != 1 {
		t.Fatal("Any must short circuit")
	}
	calls = 0
	if containers.All(seq, func(k, v int) bool { calls++; return false }) || calls != 1 {
		t.Fatal("All must short circuit")
	}
	empty := slices.All([]int(nil))
	positive := func(k, v int) bool { return v > 0 }
	if containers.Any(empty, positive) || !containers.All(empty, positive) {
		t.Fatal("incorrect empty quantifiers")
	}
	var pairs [][2]int
	filtered := containers.Filter(seq, positive)
	containers.Each(filtered, func(k, v int) { pairs = append(pairs, [2]int{k, v}) })
	if !slices.Equal(pairs, [][2]int{{1, 2}, {2, 4}}) {
		t.Fatal("filter must preserve keys")
	}
}

func TestLazyEarlyStop(t *testing.T) {
	calls := 0
	seq := func(yield func(int, int) bool) {
		for i := range 5 {
			calls++
			if !yield(i, i) {
				return
			}
		}
	}
	mapped := containers.Map(seq, func(k, v int) string { return fmt.Sprint(v) })
	filtered := containers.Filter(seq, func(k, v int) bool { return true })
	if calls != 0 {
		t.Fatal("transform must be lazy")
	}
	for range mapped {
		break
	}
	if calls != 1 {
		t.Fatal("Map did not propagate early stop")
	}
	calls = 0
	for range filtered {
		break
	}
	if calls != 1 {
		t.Fatal("Filter did not propagate early stop")
	}
}

type cursor struct {
	values []int
	index  int
}

func (c *cursor) Next() bool {
	if c.index < len(c.values) {
		c.index++
	}
	return c.index < len(c.values)
}
func (c *cursor) Value() int  { return c.values[c.index] }
func (c *cursor) Index() int  { return c.index }
func (c *cursor) Key() int    { return c.index }
func (c *cursor) Begin()      { c.index = -1 }
func (c *cursor) First() bool { c.Begin(); return c.Next() }
func (c *cursor) NextTo(p func(int, int) bool) bool {
	if p == nil {
		panic("nil predicate")
	}
	for c.Next() {
		if p(c.Index(), c.Value()) {
			return true
		}
	}
	return false
}

func TestIteratorAdapters(t *testing.T) {
	adapters := []func(*cursor) iter.Seq2[int, int]{
		func(c *cursor) iter.Seq2[int, int] { return containers.Indexed[int](c) },
		func(c *cursor) iter.Seq2[int, int] { return containers.Keyed[int, int](c) },
	}
	for _, adapt := range adapters {
		c := &cursor{values: []int{10, 20, 30}, index: -1}
		seq := adapt(c)
		for k, v := range seq {
			if k != 0 || v != 10 {
				t.Fatal("wrong first item")
			}
			break
		}
		var got []int
		for _, v := range seq {
			got = append(got, v)
		}
		if !slices.Equal(got, []int{20, 30}) {
			t.Fatal("cursor was reset", got)
		}
		for range seq {
			t.Fatal("exhausted iterator yielded")
		}
		c.Begin()
		if !containers.Any(seq, func(k, v int) bool { return v == 10 }) {
			t.Fatal("explicit reset failed")
		}
	}
}

func TestNilCallbacks(t *testing.T) {
	seq := slices.All([]int(nil))
	predicate := func(k, v int) bool { return true }
	cases := []func(){
		func() { containers.GetSortedValuesFunc[int](&source[int]{}, nil) },
		func() { containers.Each(seq, nil) },
		func() { containers.Each[int, int](nil, func(k, v int) {}) },
		func() { containers.Find(seq, nil) },
		func() { containers.Any[int, int](nil, predicate) },
		func() { containers.All(seq, nil) },
		func() { containers.Filter(seq, nil) },
		func() { containers.Map[int, int, int](seq, nil) },
		func() { containers.Map[int, int, int](nil, func(k, v int) int { return v }) },
	}
	for i, run := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			run()
		})
	}
}

func FuzzSortedSnapshot(f *testing.F) {
	f.Add([]byte{3, 1, 2})
	f.Fuzz(func(t *testing.T, values []byte) {
		before := slices.Clone(values)
		got := containers.GetSortedValues[byte](&source[byte]{values})
		if !slices.Equal(values, before) || !slices.IsSorted(got) {
			t.Fatal("sort mutated source or failed to order values")
		}
		slices.Sort(before)
		if !slices.Equal(before, got) {
			t.Fatal("sort changed membership")
		}
	})
}
