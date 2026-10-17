package trees_test

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/trees"
	"github.com/Shenghui-Waa/go.repo/trees/avltree"
	"github.com/Shenghui-Waa/go.repo/trees/binaryheap"
	"github.com/Shenghui-Waa/go.repo/trees/btree"
	"github.com/Shenghui-Waa/go.repo/trees/redblacktree"
)

type searchTree interface {
	trees.Tree[int]
	Clear()
	Put(int, int)
	Get(int) (int, bool)
	Remove(int) bool
	Keys() []int
	Floor(int) (int, int, bool)
	Ceiling(int) (int, int, bool)
	Min() (int, int, bool)
	Max() (int, int, bool)
	Iterator() *avltree.Iterator[int, int]
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}

func factories() []func() searchTree {
	return []func() searchTree{
		func() searchTree { return avltree.NewWithComparator[int, int](compare.Ordered[int]()) },
		func() searchTree { return redblacktree.NewWithComparator[int, int](compare.Ordered[int]()) },
		func() searchTree { return btree.NewWithOrderAndComparator[int, int](3, compare.Ordered[int]()) },
		func() searchTree { return btree.NewWithComparator[int, int](compare.Ordered[int]()) },
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
func exercise(t *testing.T, operations []byte) {
	for _, create := range factories() {
		tree := create()
		model := map[int]int{}
		for i, operation := range operations {
			key := int(operation) % 70
			if operation%3 == 0 {
				delete(model, key)
				tree.Remove(key)
			} else {
				model[key] = i
				tree.Put(key, i)
			}
			keys := make([]int, 0, len(model))
			for key, value := range model {
				keys = append(keys, key)
				if got, ok := tree.Get(key); !ok || got != value {
					t.Fatal("get")
				}
			}
			slices.Sort(keys)
			if !slices.Equal(keys, tree.Keys()) || len(keys) != tree.Size() {
				t.Fatal("order/size")
			}
			probe := int(operation)%75 - 2
			floor, ceiling := -1, -1
			for _, k := range keys {
				if k <= probe {
					floor = k
				}
				if k >= probe && ceiling == -1 {
					ceiling = k
				}
			}
			if k, _, ok := tree.Floor(probe); ok != (floor != -1) || (ok && k != floor) {
				t.Fatal("floor", probe, k, floor)
			}
			if k, _, ok := tree.Ceiling(probe); ok != (ceiling != -1) || (ok && k != ceiling) {
				t.Fatal("ceiling")
			}
		}
		for _, k := range tree.Keys() {
			if !tree.Remove(k) {
				t.Fatal("delete")
			}
		}
		if !tree.Empty() {
			t.Fatal("delete all")
		}
	}
}
func TestRandomTrees(t *testing.T) {
	data := make([]byte, 2500)
	rand.New(rand.NewSource(42)).Read(data)
	exercise(t, data)
}
func TestTreeContracts(t *testing.T) {
	for _, create := range factories() {
		tree := create()
		if _, _, ok := tree.Min(); ok {
			t.Fatal("empty min")
		}
		tree.Put(3, 30)
		tree.Put(1, 10)
		tree.Put(2, 20)
		if k, _, _ := tree.Min(); k != 1 {
			t.Fatal("min")
		}
		if k, _, _ := tree.Max(); k != 3 {
			t.Fatal("max")
		}
		it := tree.Iterator()
		mustPanic(t, func() { it.Key() })
		for k := 1; k <= 3; k++ {
			if !it.Next() || it.Key() != k {
				t.Fatal("forward")
			}
		}
		if it.Next() {
			t.Fatal("end")
		}
		for k := 3; k >= 1; k-- {
			if !it.Prev() || it.Key() != k {
				t.Fatal("reverse")
			}
		}
		if it.Prev() {
			t.Fatal("begin")
		}
		if !it.First() || !it.Last() {
			t.Fatal("endpoints")
		}
		before := tree.Keys()
		if tree.UnmarshalJSON([]byte(`[{"key":1,"value":"bad"}]`)) == nil {
			t.Fatal("JSON error")
		}
		if !slices.Equal(before, tree.Keys()) {
			t.Fatal("decode mutated source")
		}
		data, err := json.Marshal(tree)
		if err != nil {
			t.Fatal(err)
		}
		other := create()
		if err := json.Unmarshal(data, other); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(tree.Values(), other.Values()) {
			t.Fatal("roundtrip")
		}
		tree.Put(1, 99)
		mustPanic(t, func() { it.Begin() })
		tree.Clear()
		if !tree.Empty() {
			t.Fatal("clear")
		}
	}
}

type natural int

func (n natural) CompareTo(other natural) int { return compare.Ordered[natural]()(n, other) }
func TestConstruction(t *testing.T) {
	avltree.New[natural, int]().Put(1, 2)
	redblacktree.New[natural, int]().Put(1, 2)
	btree.New[natural, int]().Put(1, 2)
	btree.NewWithOrder[natural, int](5).Put(1, 2)
	binaryheap.New[natural]().Push(3, 1)
	mustPanic(t, func() { btree.NewWithOrder[natural, int](2) })
	mustPanic(t, func() { avltree.NewWithComparator[int, int](nil) })
	mustPanic(t, func() { binaryheap.NewWithComparator[int](nil) })
	m := redblacktree.NewWithComparator[string, int](func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	m.Put("A", 1)
	m.Put("a", 2)
	if m.Size() != 1 || m.Keys()[0] != "A" {
		t.Fatal("canonical key")
	}
	// Comparator-only objects do not need Equal or HashCode.
	type key struct{ data []int }
	custom := avltree.NewWithComparator[key, int](func(a, b key) int {
		return compare.Ordered[int]()(a.data[0], b.data[0])
	})
	custom.Put(key{[]int{1}}, 2)
	if value, ok := custom.Get(key{[]int{1}}); !ok || value != 2 {
		t.Fatal("noncomparable key")
	}
}
func TestHeap(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		order := compare.Ordered[int]()
		if reverse {
			order = order.Reversed()
		}
		h := binaryheap.NewWithComparator(order)
		rng := rand.New(rand.NewSource(81))
		var want []int
		for range 1000 {
			v := rng.Intn(100)
			want = append(want, v)
			h.Push(v)
		}
		h.Push(7, 3, 4)
		want = append(want, 7, 3, 4)
		layout := h.Values()
		for i := 1; i < len(layout); i++ {
			if order(layout[(i-1)/2], layout[i]) > 0 {
				t.Fatal("heap property")
			}
		}
		slices.SortFunc(want, order)
		for _, expected := range want {
			got, ok := h.Pop()
			if !ok || got != expected {
				t.Fatal("pop order")
			}
		}
		if _, ok := h.Pop(); ok {
			t.Fatal("empty pop")
		}
		h.Push(3, 1, 2)
		it := h.Iterator()
		mustPanic(t, func() { it.Value() })
		if !it.First() {
			t.Fatal("heap iterator")
		}
		h.Push(9)
		mustPanic(t, func() { it.Begin() })
		before := h.Values()
		if h.FromJSON([]byte(`[1,"bad"]`)) == nil {
			t.Fatal("decode")
		}
		if !slices.Equal(before, h.Values()) {
			t.Fatal("atomic heap decode")
		}
		if err := h.FromJSON([]byte(`[8,1,4]`)); err != nil {
			t.Fatal(err)
		}
		mustPanic(t, func() { h.Any(func(i, v int) bool { h.Clear(); return true }) })
	}
}
func FuzzTrees(f *testing.F) {
	f.Add([]byte{1, 4, 7, 3, 6, 9, 2, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 500 {
			data = data[:500]
		}
		exercise(t, data)
	})
}
