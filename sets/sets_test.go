package sets_test

import (
	"encoding/json"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/Shenghui-Waa/go.repo/base"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/sets"
	"github.com/Shenghui-Waa/go.repo/sets/hashset"
	"github.com/Shenghui-Waa/go.repo/sets/linkedhashset"
	"github.com/Shenghui-Waa/go.repo/sets/treeset"
)

type testSet interface {
	sets.ExtendedSet[int]
	Iterator() *hashset.Iterator[int]
	IsSubsetOf(sets.Set[int]) bool
	IsSupersetOf(sets.Set[int]) bool
	IsDisjoint(sets.Set[int]) bool
}

func factories() map[string]func() testSet {
	return map[string]func() testSet{
		"hash":   func() testSet { return hashset.New[int]() },
		"linked": func() testSet { return linkedhashset.New[int]() },
		"tree":   func() testSet { return treeset.NewWithComparator(compare.Ordered[int]()) },
	}
}
func mustPanic(t *testing.T, run func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	run()
}
func TestSetContracts(t *testing.T) {
	for name, create := range factories() {
		t.Run(name, func(t *testing.T) {
			s := create()
			if !s.Empty() || !s.Contains() || s.Add() != 0 || s.Remove() != 0 {
				t.Fatal("empty contract")
			}
			if s.Add(3, 1, 3, 2) != 3 || s.Size() != 3 {
				t.Fatal("deduplication")
			}
			if !s.Contains(1, 2) || s.Contains(4) {
				t.Fatal("membership")
			}
			values := s.Values()
			values[0] = 100
			if s.Contains(100) {
				t.Fatal("snapshot aliases source")
			}
			cursor := s.Iterator()
			var _ containers.ReverseIteratorWithIndex[int] = cursor
			mustPanic(t, func() { cursor.Value() })
			s.Add(1, 2)
			s.Remove(99)
			if !cursor.First() {
				t.Fatal("no-op invalidated cursor")
			}
			cursor.Begin()
			var got []int
			for cursor.Next() {
				got = append(got, cursor.Value())
			}
			if !slices.Equal(got, s.Values()) || cursor.Next() {
				t.Fatal("forward")
			}
			for i := len(got) - 1; i >= 0; i-- {
				if !cursor.Prev() || cursor.Index() != i || cursor.Value() != got[i] {
					t.Fatal("reverse")
				}
			}
			if cursor.Prev() || !cursor.Last() {
				t.Fatal("reverse bounds")
			}
			cursor.Begin()
			if !cursor.NextTo(func(i, v int) bool { return v == 2 }) {
				t.Fatal("next to")
			}
			cursor.End()
			if !cursor.PrevTo(func(i, v int) bool { return v == 1 }) {
				t.Fatal("prev to")
			}
			mustPanic(t, func() { cursor.NextTo(nil) })
			mustPanic(t, func() { cursor.PrevTo(nil) })
			calls := 0
			if !s.Any(func(i, v int) bool { calls++; return true }) || calls != 1 {
				t.Fatal("short circuit")
			}
			if !s.All(func(i, v int) bool { return v > 0 }) {
				t.Fatal("all")
			}
			if i, v, ok := s.Find(func(i, v int) bool { return v == 2 }); !ok || i < 0 || v != 2 {
				t.Fatal("find")
			}
			if i, _, ok := s.Find(func(i, v int) bool { return false }); ok || i != -1 {
				t.Fatal("find miss")
			}
			before := s.Values()
			if s.FromJSON([]byte(`[1,"bad"]`)) == nil || !slices.Equal(before, s.Values()) {
				t.Fatal("decode failure must be atomic")
			}
			data, err := json.Marshal(s)
			if err != nil {
				t.Fatal(err)
			}
			other := create()
			if err := json.Unmarshal(data, other); err != nil {
				t.Fatal(err)
			}
			if !other.Contains(before...) || other.Size() != len(before) {
				t.Fatal("roundtrip")
			}
			if err := s.FromJSON([]byte(`[4,4,5]`)); err != nil {
				t.Fatal(err)
			}
			if s.Size() != 2 || !s.Contains(4, 5) {
				t.Fatal("JSON deduplication")
			}
			mustPanic(t, func() { cursor.Begin() })
			mustPanic(t, func() { s.Any(func(i, v int) bool { s.Remove(v); return true }) })
			if err := s.FromJSON([]byte(`null`)); err != nil || !s.Empty() {
				t.Fatal("null")
			}
			data, _ = s.ToJSON()
			if string(data) != "[]" {
				t.Fatal("empty JSON")
			}
			if s.Any(func(i, v int) bool { return true }) || !s.All(func(i, v int) bool { return false }) {
				t.Fatal("empty quantifiers")
			}
			mustPanic(t, func() { s.Each(nil) })
			s.Add(1, 2)
			if s.Remove(1, 1, 99) != 1 {
				t.Fatal("remove count")
			}
			s.Clear()
			if !s.Empty() {
				t.Fatal("clear")
			}
		})
	}
}

func TestAlgebraAndTransforms(t *testing.T) {
	a := hashset.New[int]()
	b := linkedhashset.New[int]()
	c := treeset.NewWithComparator(compare.Ordered[int]().Reversed())
	other := hashset.New[int]()
	other.Add(2, 4)
	for _, s := range []sets.Set[int]{a, b, c} {
		s.Add(3, 1, 2)
	}
	checks := []struct {
		result sets.ExtendedSet[int]
		want   []int
	}{
		{a.Union(other), []int{1, 2, 3, 4}},
		{b.Union(other), []int{1, 2, 3, 4}},
		{c.Union(other), []int{1, 2, 3, 4}},
		{a.Intersection(other), []int{2}},
		{b.Intersection(other), []int{2}},
		{c.Intersection(other), []int{2}},
		{a.Difference(other), []int{1, 3}},
		{b.Difference(other), []int{1, 3}},
		{c.Difference(other), []int{1, 3}},
		{a.SymmetricDifference(other), []int{1, 3, 4}},
		{b.SymmetricDifference(other), []int{1, 3, 4}},
		{c.SymmetricDifference(other), []int{1, 3, 4}},
	}
	for _, check := range checks {
		values := check.result.Values()
		slices.Sort(values)
		if !slices.Equal(values, check.want) {
			t.Fatal("algebra", values, check.want)
		}
	}
	if !slices.Equal(b.Union(other).Values(), []int{3, 1, 2, 4}) {
		t.Fatal("union order")
	}
	if !slices.Equal(c.Union(other).Values(), []int{4, 3, 2, 1}) {
		t.Fatal("comparator lost")
	}
	if a.Size() != 3 || other.Size() != 2 {
		t.Fatal("source mutated")
	}
	for _, s := range []testSet{a, b, c} {
		if !s.IsSubsetOf(s) || !s.IsSupersetOf(s) || s.IsDisjoint(s) {
			t.Fatal("self relation")
		}
		if s.IsSubsetOf(other) || s.IsSupersetOf(other) {
			t.Fatal("relation")
		}
		if !s.IsDisjoint(hashset.New[int]()) {
			t.Fatal("empty disjoint")
		}
	}
	transform := func(i, v int) int { return v % 2 }
	predicate := func(i, v int) bool { return v > 1 }
	for _, result := range []sets.Set[int]{a.Map(transform), b.Map(transform), c.Map(transform)} {
		if result.Size() != 2 || !result.Contains(0, 1) {
			t.Fatal("map deduplication")
		}
	}
	for _, result := range []sets.Set[int]{a.Select(predicate), b.Select(predicate), c.Select(predicate)} {
		if result.Size() != 2 || !result.Contains(2, 3) {
			t.Fatal("select")
		}
	}
	for _, run := range []func(){
		func() { a.Map(nil) }, func() { b.Map(nil) }, func() { c.Map(nil) },
		func() { a.Select(nil) }, func() { b.Select(nil) }, func() { c.Select(nil) },
		func() { a.Union(nil) }, func() { var nilSet *hashset.Set[int]; c.Difference(nilSet) },
	} {
		mustPanic(t, run)
	}
}

type collision struct {
	ID      int
	Payload []int
}

func (v collision) Equal(other any) bool {
	value, ok := other.(collision)
	return ok && v.ID == value.ID
}
func (collision) HashCode() uint64 { return 1 }

type identity struct{ base.Base }
type natural int

func (n natural) CompareTo(other natural) int { return compare.Ordered[natural]()(n, other) }

func TestMembershipPolicies(t *testing.T) {
	s := hashset.New[collision]()
	for i := range 40 {
		s.Add(collision{ID: i})
	}
	if s.Add(collision{ID: 3, Payload: []int{1}}) != 0 || s.Size() != 40 {
		t.Fatal("collision dedup")
	}
	for i := range 40 {
		if !s.Contains(collision{ID: i}) {
			t.Fatal("collision lookup")
		}
	}
	objects := linkedhashset.New[*identity]()
	a, b := &identity{}, &identity{}
	if objects.Add(a, a, b, nil) != 3 || !objects.Contains(a, nil) {
		t.Fatal("Base identity")
	}
	bad := hashset.New[any]()
	mustPanic(t, func() { bad.Add(1, struct{}{}) })
	if !bad.Empty() {
		t.Fatal("invalid batch partially inserted")
	}
	bad.Add("kept")
	if bad.FromJSON([]byte(`[1,{}]`)) == nil || !bad.Contains("kept") || bad.Size() != 1 {
		t.Fatal("unsupported JSON key rollback")
	}
	n := treeset.New[natural]()
	n.Add(3, 1, 2)
	if v, ok := n.Min(); !ok || v != 1 {
		t.Fatal("min")
	}
	if v, ok := n.Max(); !ok || v != 3 {
		t.Fatal("max")
	}
	if v, ok := n.Floor(2); !ok || v != 2 {
		t.Fatal("floor")
	}
	if v, ok := n.Ceiling(2); !ok || v != 2 {
		t.Fatal("ceiling")
	}
	if _, ok := n.Floor(0); ok {
		t.Fatal("floor miss")
	}
	mustPanic(t, func() { treeset.NewWithComparator[int](nil) })
	caseInsensitive := treeset.NewWithComparator(compare.Comparator[string](func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	}))
	caseInsensitive.Add("A")
	caseSensitive := hashset.New[string]()
	caseSensitive.Add("a", "B")
	if caseInsensitive.Intersection(caseSensitive).Size() != 1 {
		t.Fatal("receiver equivalence")
	}
	if caseSensitive.Intersection(caseInsensitive).Size() != 0 {
		t.Fatal("hash equivalence")
	}
	if caseInsensitive.Add("a") != 0 || caseInsensitive.Values()[0] != "A" {
		t.Fatal("representative")
	}
	var zero linkedhashset.Set[int]
	zero.Add(2, 1)
	zero.Remove(2)
	zero.Add(2)
	if !slices.Equal(zero.Values(), []int{1, 2}) {
		t.Fatal("zero/reinsert order")
	}
}

func runModel(t *testing.T, operations []byte) {
	for _, create := range factories() {
		s := create()
		model := map[int]bool{}
		for _, op := range operations {
			value := int(op) % 40
			if op%3 == 0 {
				s.Remove(value)
				delete(model, value)
			} else {
				s.Add(value)
				model[value] = true
			}
			if s.Size() != len(model) {
				t.Fatal("model size")
			}
			for _, value := range s.Values() {
				if !model[value] {
					t.Fatal("unexpected member")
				}
			}
			for value := range model {
				if !s.Contains(value) {
					t.Fatal("missing member")
				}
			}
		}
	}
}
func TestRandomModel(t *testing.T) {
	data := make([]byte, 3000)
	rand.New(rand.NewSource(42)).Read(data)
	runModel(t, data)
}
func FuzzSets(f *testing.F) {
	f.Add([]byte{1, 2, 3, 1, 6, 7, 8, 9})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 500 {
			data = data[:500]
		}
		runModel(t, data)
	})
}
