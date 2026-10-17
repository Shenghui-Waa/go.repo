package stacks_test

import (
	"encoding/json"
	"math/rand"
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/stacks"
	"github.com/Shenghui-Waa/go.repo/stacks/arraystack"
	"github.com/Shenghui-Waa/go.repo/stacks/linkedliststack"
)

type testStack interface {
	stacks.ExtendedStack[int]
	Iterator() *arraystack.Iterator[int]
}

func factories() map[string]func() testStack {
	return map[string]func() testStack{
		"array":  func() testStack { return arraystack.New[int]() },
		"linked": func() testStack { return linkedliststack.New[int]() },
	}
}
func mustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	f()
}
func TestContracts(t *testing.T) {
	for name, create := range factories() {
		t.Run(name, func(t *testing.T) {
			s := create()
			if !s.Empty() {
				t.Fatal("empty")
			}
			if _, ok := s.Pop(); ok {
				t.Fatal("empty pop")
			}
			if _, ok := s.Peek(); ok {
				t.Fatal("empty peek")
			}
			idle := s.Iterator()
			s.Clear()
			s.Pop()
			if idle.Next() {
				t.Fatal("empty iterator")
			}
			for _, v := range []int{1, 2, 2, 3} {
				s.Push(v)
			}
			if !slices.Equal(s.Values(), []int{3, 2, 2, 1}) {
				t.Fatal("top-first")
			}
			values := s.Values()
			values[0] = 99
			if v, _ := s.Peek(); v != 3 {
				t.Fatal("snapshot alias")
			}
			it := s.Iterator()
			var _ containers.ReverseIteratorWithIndex[int] = it
			mustPanic(t, func() { it.Value() })
			var got []int
			for it.Next() {
				got = append(got, it.Value())
			}
			if !slices.Equal(got, s.Values()) || it.Next() {
				t.Fatal("forward")
			}
			for i := len(got) - 1; i >= 0; i-- {
				if !it.Prev() || it.Index() != i || it.Value() != got[i] {
					t.Fatal("reverse")
				}
			}
			if it.Prev() || !it.First() || !it.Last() {
				t.Fatal("bounds")
			}
			it.Begin()
			if !it.NextTo(func(i, v int) bool { return v == 2 }) {
				t.Fatal("next to")
			}
			it.End()
			if !it.PrevTo(func(i, v int) bool { return v == 3 }) {
				t.Fatal("prev to")
			}
			mustPanic(t, func() { it.NextTo(nil) })
			mustPanic(t, func() { it.PrevTo(nil) })
			calls := 0
			if !s.Any(func(i, v int) bool { calls++; return true }) || calls != 1 {
				t.Fatal("short circuit")
			}
			if !s.All(func(i, v int) bool { return v > 0 }) {
				t.Fatal("all")
			}
			if i, _, ok := s.Find(func(i, v int) bool { return false }); ok || i != -1 {
				t.Fatal("find miss")
			}
			if i, v, ok := s.Find(func(i, v int) bool { return v == 3 }); !ok || i != 0 || v != 3 {
				t.Fatal("find")
			}
			before := s.Values()
			if s.FromJSON([]byte(`[9,"bad"]`)) == nil || !slices.Equal(before, s.Values()) {
				t.Fatal("decode rollback")
			}
			data, err := json.Marshal(s)
			if err != nil || string(data) != "[3,2,2,1]" {
				t.Fatal("JSON order", string(data), err)
			}
			other := create()
			if err := json.Unmarshal(data, other); err != nil {
				t.Fatal(err)
			}
			for !s.Empty() {
				a, _ := s.Pop()
				b, ok := other.Pop()
				if !ok || a != b {
					t.Fatal("roundtrip reversed stack")
				}
			}
			mustPanic(t, func() { it.Begin() })
			s.Push(1)
			seq := s.Seq()
			s.Push(2)
			mustPanic(t, func() {
				for range seq {
				}
			})
			mustPanic(t, func() { s.Any(func(i, v int) bool { s.Pop(); return true }) })
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
			s.Push(1)
			s.Clear()
			if !s.Empty() {
				t.Fatal("clear")
			}
		})
	}
}
func TestTransformOrderAndArbitraryTypes(t *testing.T) {
	a, b := arraystack.New[int](), linkedliststack.New[int]()
	for _, s := range []stacks.ExtendedStack[int]{a, b} {
		s.Push(1)
		s.Push(2)
		s.Push(3)
	}
	f := func(i, v int) int { return i*10 + v }
	p := func(i, v int) bool { return v != 2 }
	for _, s := range []stacks.ExtendedStack[int]{a.Map(f), b.Map(f)} {
		if !slices.Equal(s.Values(), []int{3, 12, 21}) {
			t.Fatal("map reversed")
		}
	}
	for _, s := range []stacks.ExtendedStack[int]{a.Select(p), b.Select(p)} {
		if !slices.Equal(s.Values(), []int{3, 1}) {
			t.Fatal("select reversed")
		}
	}
	for _, run := range []func(){
		func() { a.Map(nil) }, func() { b.Map(nil) },
		func() { a.Select(nil) }, func() { b.Select(nil) },
	} {
		mustPanic(t, run)
	}
	var zero arraystack.Stack[[]int]
	zero.Push([]int{1})
	zero.Push(nil)
	if v, ok := zero.Pop(); !ok || v != nil {
		t.Fatal("nil slice")
	}
	if v, ok := zero.Pop(); !ok || v[0] != 1 {
		t.Fatal("slice")
	}
	var linked linkedliststack.Stack[map[string]int]
	linked.Push(map[string]int{"x": 1})
	if v, ok := linked.Pop(); !ok || v["x"] != 1 {
		t.Fatal("map value")
	}
}
func runModel(t *testing.T, data []byte) {
	for _, create := range factories() {
		s := create()
		var model []int
		for _, op := range data {
			if op%3 == 0 {
				value, ok := s.Pop()
				if ok != (len(model) > 0) {
					t.Fatal("pop status")
				}
				if ok {
					if value != model[len(model)-1] {
						t.Fatal("pop order")
					}
					model = model[:len(model)-1]
				}
			} else {
				model = append(model, int(op))
				s.Push(int(op))
			}
			if s.Size() != len(model) {
				t.Fatal("size")
			}
			if value, ok := s.Peek(); ok != (len(model) > 0) || (ok && value != model[len(model)-1]) {
				t.Fatal("peek")
			}
			values := s.Values()
			slices.Reverse(values)
			if !slices.Equal(values, model) {
				t.Fatal("snapshot order")
			}
		}
	}
}
func TestRandomModel(t *testing.T) {
	data := make([]byte, 3000)
	rand.New(rand.NewSource(42)).Read(data)
	runModel(t, data)
}
func FuzzStacks(f *testing.F) {
	f.Add([]byte{1, 2, 3, 3, 4, 5, 6})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 500 {
			data = data[:500]
		}
		runModel(t, data)
	})
}
