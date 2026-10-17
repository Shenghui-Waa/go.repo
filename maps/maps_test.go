package maps_test

import (
	"encoding/json"
	"math"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Shenghui-Waa/go.repo/base"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/maps"
	"github.com/Shenghui-Waa/go.repo/maps/hashbidimap"
	"github.com/Shenghui-Waa/go.repo/maps/hashmap"
	"github.com/Shenghui-Waa/go.repo/maps/linkedhashmap"
	"github.com/Shenghui-Waa/go.repo/maps/treebidimap"
	"github.com/Shenghui-Waa/go.repo/maps/treemap"
)

type testMap interface {
	maps.ExtendedMap[int, int]
	Iterator() *hashmap.Iterator[int, int]
}

func factories() map[string]func() testMap {
	return map[string]func() testMap{
		"hash":     func() testMap { return hashmap.New[int, int]() },
		"linked":   func() testMap { return linkedhashmap.New[int, int]() },
		"tree":     func() testMap { return treemap.NewWithComparator[int, int](compare.Ordered[int]()) },
		"hashbidi": func() testMap { return hashbidimap.New[int, int]() },
		"treebidi": func() testMap {
			return treebidimap.NewWithComparator(compare.Ordered[int](), compare.Ordered[int]())
		},
	}
}
func panicExpected(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	f()
}

func TestCommon(t *testing.T) {
	for name, create := range factories() {
		t.Run(name, func(t *testing.T) {
			m := create()
			if !m.Empty() || m.Remove(0) {
				t.Fatal("empty")
			}
			m.Put(3, 30)
			m.Put(1, 10)
			m.Put(2, 20)
			m.Put(1, 11)
			if m.Size() != 3 || !m.ContainsKey(1) || !m.ContainsValue(11) {
				t.Fatal("put/update")
			}
			if m.ContainsValue(99) {
				t.Fatal("missing value")
			}
			if value, ok := m.Get(99); ok || value != 0 {
				t.Fatal("missing get")
			}
			entries := m.Entries()
			entries[0].Value = -1
			if m.ContainsValue(-1) {
				t.Fatal("snapshot alias")
			}
			if !m.All(func(k, v int) bool { return v > 0 }) {
				t.Fatal("all")
			}
			calls := 0
			if !m.Any(func(k, v int) bool { calls++; return true }) || calls != 1 {
				t.Fatal("short circuit")
			}
			if k, v, ok := m.Find(func(k, v int) bool { return k == 2 }); !ok || k != 2 || v != 20 {
				t.Fatal("find")
			}
			if _, _, ok := m.Find(func(k, v int) bool { return false }); ok {
				t.Fatal("find miss")
			}
			c := m.Iterator()
			var _ containers.ReverseIteratorWithKey[int, int] = c
			panicExpected(t, func() { c.Value() })
			var keys []int
			for c.Next() {
				keys = append(keys, c.Key())
			}
			if !slices.Equal(keys, m.Keys()) || c.Next() {
				t.Fatal("iterator")
			}
			for i := len(keys) - 1; i >= 0; i-- {
				if !c.Prev() || c.Key() != keys[i] {
					t.Fatal("reverse")
				}
			}
			if c.Prev() {
				t.Fatal("reverse end")
			}
			if !c.First() || !c.Last() {
				t.Fatal("endpoints")
			}
			c.Begin()
			if !c.NextTo(func(k, v int) bool { return k == 2 }) {
				t.Fatal("next to")
			}
			c.End()
			if !c.PrevTo(func(k, v int) bool { return k == 1 }) {
				t.Fatal("prev to")
			}
			panicExpected(t, func() { c.NextTo(nil) })
			panicExpected(t, func() { c.PrevTo(nil) })
			before := m.Entries()
			if m.FromJSON([]byte(`[{"key":1,"value":"bad"}]`)) == nil {
				t.Fatal("invalid decode accepted")
			}
			if !reflect.DeepEqual(before, m.Entries()) {
				t.Fatal("decode not atomic")
			}
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			n := create()
			if err := json.Unmarshal(data, n); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(m.Entries(), n.Entries()) {
				t.Fatal("roundtrip")
			}
			m.Put(1, 12)
			panicExpected(t, func() { c.Begin() })
			panicExpected(t, func() { m.Any(func(k, v int) bool { m.Remove(k); return true }) })
			if err := m.FromJSON([]byte(`null`)); err != nil || !m.Empty() {
				t.Fatal("null")
			}
			m.Put(0, 0)
			if _, _, ok := m.Find(func(k, v int) bool { return true }); !ok {
				t.Fatal("zero match")
			}
			m.Clear()
			data, _ = m.ToJSON()
			if string(data) != "[]" || !m.Empty() {
				t.Fatal("empty JSON")
			}
			panicExpected(t, func() { m.Each(nil) })
		})
	}
}

type collision struct {
	ID      int
	Payload []int
}

func (c collision) Equal(other any) bool { v, ok := other.(collision); return ok && c.ID == v.ID }
func (collision) HashCode() uint64       { return 7 }

type identity struct {
	base.Base
	Name string
}
type natural int

func (n natural) CompareTo(other natural) int { return compare.Ordered[natural]()(n, other) }

func TestHashKeys(t *testing.T) {
	m := hashmap.New[collision, []int]()
	for i := range 50 {
		m.Put(collision{ID: i}, []int{i})
	}
	for i := range 50 {
		v, ok := m.Get(collision{ID: i, Payload: []int{1}})
		if !ok || v[0] != i {
			t.Fatal("collision lookup")
		}
	}
	for i := range 50 {
		if !m.Remove(collision{ID: i}) {
			t.Fatal("collision removal")
		}
	}
	if !m.Empty() {
		t.Fatal("collision size")
	}
	n := hashmap.New[*identity, int]()
	a, b := &identity{Name: "x"}, &identity{Name: "x"}
	n.Put(a, 1)
	if _, ok := n.Get(a); !ok {
		t.Fatal("embedded Base")
	}
	if n.ContainsKey(b) {
		t.Fatal("distinct identity")
	}
	n.Put(nil, 2)
	if v, ok := n.Get(nil); !ok || v != 2 {
		t.Fatal("nil key")
	}
	panicExpected(t, func() { hashmap.New[float64, int]().Put(math.NaN(), 1) })
	panicExpected(t, func() { hashmap.New[struct{}, int]().Put(struct{}{}, 1) })
	anyMap := hashmap.New[any, any]()
	anyMap.Put("kept", []int{1})
	if anyMap.FromJSON([]byte(`[{"key":{"x":1},"value":2}]`)) == nil {
		t.Fatal("unsupported key should return decoding error")
	}
	if !anyMap.ContainsKey("kept") {
		t.Fatal("contract failure lost source")
	}
}

func TestTreeSemantics(t *testing.T) {
	m := treemap.NewWithComparator[string, int](func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	m.Put("A", 1)
	m.Put("a", 2)
	if m.Size() != 1 || m.Keys()[0] != "A" {
		t.Fatal("comparator identity/canonical key")
	}
	if v, ok := m.Get("a"); !ok || v != 2 {
		t.Fatal("equivalent lookup")
	}
	n := treemap.New[natural, string]()
	for _, k := range []natural{5, 1, 3} {
		n.Put(k, "v")
	}
	if k, _, ok := n.Min(); !ok || k != 1 {
		t.Fatal("min")
	}
	if k, _, ok := n.Max(); !ok || k != 5 {
		t.Fatal("max")
	}
	if k, _, ok := n.Floor(4); !ok || k != 3 {
		t.Fatal("floor")
	}
	if k, _, ok := n.Ceiling(4); !ok || k != 5 {
		t.Fatal("ceiling")
	}
	if _, _, ok := n.Floor(0); ok {
		t.Fatal("floor miss")
	}
	if _, _, ok := n.Ceiling(6); ok {
		t.Fatal("ceiling miss")
	}
	b := treebidimap.New[natural, natural]()
	b.Put(1, 10)
	if k, ok := b.GetKey(10); !ok || k != 1 {
		t.Fatal("natural bidi")
	}
	panicExpected(t, func() { treemap.NewWithComparator[int, int](nil) })
	panicExpected(t, func() { treebidimap.NewWithComparator[int, int](compare.Ordered[int](), nil) })
	panicExpected(t, func() { var zero treemap.Map[int, int]; zero.Put(1, 1) })
}

func TestBidiConflicts(t *testing.T) {
	for _, m := range []maps.ExtendedBidiMap[int, int]{
		hashbidimap.New[int, int](),
		treebidimap.NewWithComparator(compare.Ordered[int](), compare.Ordered[int]()),
	} {
		m.Put(1, 10)
		m.Put(2, 20)
		m.Put(1, 20)
		if m.Size() != 1 || m.ContainsKey(2) {
			t.Fatal("two conflicts")
		}
		if _, ok := m.GetKey(10); ok {
			t.Fatal("stale reverse value")
		}
		if key, ok := m.GetKey(20); !ok || key != 1 {
			t.Fatal("reverse owner")
		}
		m.Put(3, 20)
		if m.ContainsKey(1) || m.Size() != 1 {
			t.Fatal("value eviction")
		}
		if !m.RemoveValue(20) || m.RemoveValue(20) || !m.Empty() {
			t.Fatal("remove value")
		}
	}
}

func TestLinkedOrderAndTransforms(t *testing.T) {
	m := linkedhashmap.New[int, int]()
	m.Put(3, 30)
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(1, 11)
	if !slices.Equal(m.Keys(), []int{3, 1, 2}) {
		t.Fatal("replacement moved key")
	}
	m.Remove(1)
	m.Put(1, 12)
	if !slices.Equal(m.Keys(), []int{3, 2, 1}) {
		t.Fatal("reinsertion order")
	}
	if m.Select(func(k, v int) bool { return k > 1 }).Size() != 2 {
		t.Fatal("select")
	}
	if m.Map(func(k, v int) (int, int) { return k + 1, v }).Size() != 3 {
		t.Fatal("map")
	}
}

func checkModel(t *testing.T, m testMap, model map[int]int) {
	t.Helper()
	if m.Size() != len(model) {
		t.Fatal("size mismatch", m.Entries(), model)
	}
	for k, v := range model {
		got, ok := m.Get(k)
		if !ok || got != v {
			t.Fatal("forward mismatch")
		}
		if bidi, ok := m.(maps.ExtendedBidiMap[int, int]); ok {
			key, found := bidi.GetKey(v)
			if !found || key != k {
				t.Fatal("inverse mismatch")
			}
		}
	}
}
func runOperations(t *testing.T, data []byte) {
	for _, create := range factories() {
		m := create()
		model := map[int]int{}
		_, bidi := m.(maps.ExtendedBidiMap[int, int])
		for i := 0; i+2 < len(data); i += 3 {
			op, k, v := data[i]%4, int(data[i+1]%40), int(data[i+2]%40)
			switch op {
			case 0, 1:
				if bidi {
					for key, value := range model {
						if value == v {
							delete(model, key)
						}
					}
				}
				model[k] = v
				m.Put(k, v)
			case 2:
				delete(model, k)
				m.Remove(k)
			case 3:
				if v == 0 {
					clear(model)
					m.Clear()
				}
			}
			checkModel(t, m, model)
		}
	}
}
func TestRandomModel(t *testing.T) {
	data := make([]byte, 9000)
	rand.New(rand.NewSource(42)).Read(data)
	runOperations(t, data)
}
func FuzzMaps(f *testing.F) {
	f.Add([]byte{0, 1, 10, 0, 2, 20, 0, 1, 20, 2, 1, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1500 {
			data = data[:1500]
		}
		runOperations(t, data)
	})
}
