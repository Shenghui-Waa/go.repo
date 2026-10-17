package queues_test

import (
	"encoding/json"
	"math/rand"
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/queues"
	"github.com/Shenghui-Waa/go.repo/queues/arrayqueue"
	"github.com/Shenghui-Waa/go.repo/queues/circularbuffer"
	"github.com/Shenghui-Waa/go.repo/queues/linkedlistqueue"
	"github.com/Shenghui-Waa/go.repo/queues/priorityqueue"
)

type testQueue interface {
	queues.ExtendedQueue[int]
	Iterator() *arrayqueue.Iterator[int]
}

func factories() map[string]func() testQueue {
	return map[string]func() testQueue{
		"array":    func() testQueue { return arrayqueue.New[int]() },
		"linked":   func() testQueue { return linkedlistqueue.New[int]() },
		"ring":     func() testQueue { return circularbuffer.NewWithCapacity[int](7) },
		"priority": func() testQueue { return priorityqueue.NewWithComparator(compare.Ordered[int]()) },
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
func TestContracts(t *testing.T) {
	for name, create := range factories() {
		t.Run(name, func(t *testing.T) {
			q := create()
			if !q.Empty() {
				t.Fatal("empty")
			}
			if _, ok := q.Dequeue(); ok {
				t.Fatal("empty dequeue")
			}
			if _, ok := q.Peek(); ok {
				t.Fatal("empty peek")
			}
			for _, v := range []int{3, 1, 2, 2} {
				q.Enqueue(v)
			}
			if q.Size() != 4 {
				t.Fatal("duplicates")
			}
			values := q.Values()
			values[0] = 99
			if v, _ := q.Peek(); v == 99 {
				t.Fatal("alias")
			}
			it := q.Iterator()
			var _ containers.ReverseIteratorWithIndex[int] = it
			mustPanic(t, func() { it.Value() })
			var got []int
			for it.Next() {
				got = append(got, it.Value())
			}
			if !slices.Equal(got, q.Values()) || it.Next() {
				t.Fatal("forward")
			}
			for i := len(got) - 1; i >= 0; i-- {
				if !it.Prev() || it.Index() != i || it.Value() != got[i] {
					t.Fatal("reverse")
				}
			}
			if it.Prev() || !it.First() || !it.Last() {
				t.Fatal("boundaries")
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
			if !q.Any(func(i, v int) bool { calls++; return true }) || calls != 1 {
				t.Fatal("short circuit")
			}
			if !q.All(func(i, v int) bool { return v > 0 }) {
				t.Fatal("all")
			}
			if i, _, ok := q.Find(func(i, v int) bool { return false }); ok || i != -1 {
				t.Fatal("find miss")
			}
			if _, v, ok := q.Find(func(i, v int) bool { return v == 1 }); !ok || v != 1 {
				t.Fatal("find")
			}
			before := q.Values()
			if q.FromJSON([]byte(`[4,"bad"]`)) == nil || !slices.Equal(q.Values(), before) {
				t.Fatal("decode rollback")
			}
			data, err := json.Marshal(q)
			if err != nil {
				t.Fatal(err)
			}
			other := create()
			if err := json.Unmarshal(data, other); err != nil {
				t.Fatal(err)
			}
			for !q.Empty() {
				a, _ := q.Dequeue()
				b, ok := other.Dequeue()
				if !ok || a != b {
					t.Fatal("roundtrip dequeue order")
				}
			}
			mustPanic(t, func() { it.Begin() })
			q.Enqueue(1)
			mustPanic(t, func() { q.Any(func(i, v int) bool { q.Enqueue(2); return true }) })
			if err := q.FromJSON([]byte(`null`)); err != nil || !q.Empty() {
				t.Fatal("null")
			}
			data, _ = q.ToJSON()
			if string(data) != "[]" {
				t.Fatal("empty JSON")
			}
			mustPanic(t, func() { q.Each(nil) })
		})
	}
}
func runModel(t *testing.T, operations []byte) {
	for name, create := range factories() {
		q := create()
		var model []int
		for _, op := range operations {
			if op%3 != 0 {
				q.Enqueue(int(op))
				model = append(model, int(op))
				if name == "priority" {
					slices.Sort(model)
				}
				if name == "ring" && len(model) > 7 {
					model = model[1:]
				}
			} else {
				v, ok := q.Dequeue()
				if ok != (len(model) > 0) {
					t.Fatal("dequeue status", name)
				}
				if ok {
					if v != model[0] {
						t.Fatal("dequeue value", name)
					}
					model = model[1:]
				}
			}
			if q.Size() != len(model) {
				t.Fatal("size", name)
			}
			if v, ok := q.Peek(); ok != (len(model) > 0) || (ok && v != model[0]) {
				t.Fatal("peek", name)
			}
			values := q.Values()
			if name == "priority" {
				slices.Sort(values)
			}
			if !slices.Equal(values, model) {
				t.Fatal("values", name)
			}
		}
	}
}
func TestRandomModel(t *testing.T) {
	data := make([]byte, 3000)
	rand.New(rand.NewSource(42)).Read(data)
	runModel(t, data)
}

type natural int

func (n natural) CompareTo(other natural) int { return compare.Ordered[natural]()(n, other) }
func TestConstructionAndCircularOverwrite(t *testing.T) {
	q := circularbuffer.NewWithCapacity[int](1)
	q.Enqueue(1)
	q.Enqueue(2)
	if !q.Full() || q.Capacity() != 1 {
		t.Fatal("capacity")
	}
	if v, _ := q.Dequeue(); v != 2 || q.Full() {
		t.Fatal("overwrite")
	}
	if err := q.FromJSON([]byte(`[1,2,3]`)); err != nil {
		t.Fatal(err)
	}
	if v, _ := q.Peek(); v != 3 || q.Capacity() != 1 {
		t.Fatal("decode overflow")
	}
	if circularbuffer.New[int]().Capacity() != circularbuffer.DefaultCapacity {
		t.Fatal("default")
	}
	mustPanic(t, func() { circularbuffer.NewWithCapacity[int](0) })
	mustPanic(t, func() { circularbuffer.NewWithCapacity[int](-1) })
	mustPanic(t, func() { priorityqueue.NewWithComparator[int](nil) })
	p := priorityqueue.New[natural]()
	p.Enqueue(3)
	p.Enqueue(1)
	if v, _ := p.Dequeue(); v != 1 {
		t.Fatal("natural")
	}
	maxQueue := priorityqueue.NewWithComparator(compare.Ordered[int]().Reversed())
	maxQueue.Enqueue(1)
	maxQueue.Enqueue(9)
	if v, _ := maxQueue.Dequeue(); v != 9 {
		t.Fatal("max priority")
	}
	var zero arrayqueue.Queue[[]int]
	zero.Enqueue([]int{1})
	if v, ok := zero.Dequeue(); !ok || v[0] != 1 {
		t.Fatal("arbitrary values")
	}
	var linked linkedlistqueue.Queue[map[string]int]
	linked.Enqueue(map[string]int{"a": 1})
	linked.Clear()
	linked.Enqueue(nil)
	if v, ok := linked.Dequeue(); !ok || v != nil {
		t.Fatal("nil value")
	}
}
func TestTransforms(t *testing.T) {
	a := arrayqueue.New[int]()
	l := linkedlistqueue.New[int]()
	c := circularbuffer.NewWithCapacity[int](5)
	p := priorityqueue.NewWithComparator(compare.Ordered[int]().Reversed())
	for _, q := range []queues.ExtendedQueue[int]{a, l, c, p} {
		q.Enqueue(1)
		q.Enqueue(2)
	}
	f := func(i, v int) int { return v + 10 }
	filter := func(i, v int) bool { return v == 2 }
	for _, q := range []queues.ExtendedQueue[int]{a.Map(f), l.Map(f), c.Map(f), p.Map(f)} {
		if q.Size() != 2 {
			t.Fatal("map")
		}
	}
	for _, q := range []queues.ExtendedQueue[int]{a.Select(filter), l.Select(filter), c.Select(filter), p.Select(filter)} {
		if v, ok := q.Dequeue(); !ok || v != 2 || !q.Empty() {
			t.Fatal("select")
		}
	}
	if c.Map(f).Capacity() != 5 {
		t.Fatal("capacity lost")
	}
	if v, _ := p.Map(f).Dequeue(); v != 12 {
		t.Fatal("comparator lost")
	}
	for _, run := range []func(){
		func() { a.Map(nil) }, func() { l.Map(nil) }, func() { c.Map(nil) }, func() { p.Map(nil) },
		func() { a.Select(nil) }, func() { l.Select(nil) }, func() { c.Select(nil) }, func() { p.Select(nil) },
	} {
		mustPanic(t, run)
	}
}
func FuzzQueues(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 500 {
			data = data[:500]
		}
		runModel(t, data)
	})
}
