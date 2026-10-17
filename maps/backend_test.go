package maps_test

import (
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/maps/treebidimap"
	"github.com/Shenghui-Waa/go.repo/maps/treemap"
)

func TestLocalBackendNonComparableKeys(t *testing.T) {
	order := compare.Comparator[[]int](func(a, b []int) int {
		return slices.Compare(a, b)
	})
	m := treemap.NewWithComparator[[]int, []string](order)
	for _, key := range []int{3, 1, 5} {
		m.Put([]int{key}, []string{"value"})
	}
	m.Put([]int{3}, []string{"updated"})
	if m.Size() != 3 {
		t.Fatal("equivalent key created another entry")
	}
	if value, ok := m.Get([]int{3}); !ok || value[0] != "updated" {
		t.Fatal("typed payload lookup failed")
	}
	if key, _, ok := m.Floor([]int{4}); !ok || key[0] != 3 {
		t.Fatal("floor failed")
	}
	if key, _, ok := m.Ceiling([]int{4}); !ok || key[0] != 5 {
		t.Fatal("ceiling failed")
	}
	if !m.Remove([]int{3}) || m.ContainsKey([]int{3}) {
		t.Fatal("non-comparable removal failed")
	}
	b := treebidimap.NewWithComparator(order, order)
	b.Put([]int{1}, []int{10})
	b.Put([]int{2}, []int{20})
	b.Put([]int{1}, []int{20})
	if b.Size() != 1 || b.ContainsKey([]int{2}) {
		t.Fatal("two-index conflict handling failed")
	}
	if key, ok := b.GetKey([]int{20}); !ok || key[0] != 1 {
		t.Fatal("reverse typed index failed")
	}
}

func TestLocalBackendNilKeysAndDescendingBounds(t *testing.T) {
	order := compare.NilFirst(compare.Comparator[*int](func(a, b *int) int {
		return compare.Ordered[int]()(*a, *b)
	}))
	m := treemap.NewWithComparator[*int, string](order)
	value := 2
	m.Put(nil, "nil")
	m.Put(&value, "two")
	if key, value, ok := m.Min(); !ok || key != nil || value != "nil" {
		t.Fatal("nil key minimum failed")
	}
	if !m.Remove(nil) || m.Size() != 1 {
		t.Fatal("nil key deletion failed")
	}
	d := treemap.NewWithComparator[int, int](compare.Ordered[int]().Reversed())
	for _, key := range []int{1, 3, 5} {
		d.Put(key, key)
	}
	if key, _, ok := d.Min(); !ok || key != 5 {
		t.Fatal("descending minimum")
	}
	if key, _, ok := d.Max(); !ok || key != 1 {
		t.Fatal("descending maximum")
	}
	if key, _, ok := d.Floor(4); !ok || key != 5 {
		t.Fatal("descending floor")
	}
	if key, _, ok := d.Ceiling(4); !ok || key != 3 {
		t.Fatal("descending ceiling")
	}
	d.Clear()
	if _, _, ok := d.Min(); ok {
		t.Fatal("empty bound")
	}
}
