package lists_test

import (
	"encoding/json"
	"math/rand"
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/base"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/lists"
	"github.com/Shenghui-Waa/go.repo/lists/arraylist"
	"github.com/Shenghui-Waa/go.repo/lists/doublylinkedlist"
	"github.com/Shenghui-Waa/go.repo/lists/singlylinkedlist"
)

func factories() map[string]func() lists.ExtendedList[int] {
	return map[string]func() lists.ExtendedList[int]{
		"array":  func() lists.ExtendedList[int] { return arraylist.New[int]() },
		"single": func() lists.ExtendedList[int] { return singlylinkedlist.New[int]() },
		"double": func() lists.ExtendedList[int] { return doublylinkedlist.New[int]() },
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

func TestListContracts(t *testing.T) {
	for name, newList := range factories() {
		t.Run(name, func(t *testing.T) {
			l := newList()
			if !l.Empty() || l.Size() != 0 || !l.Contains() {
				t.Fatal("empty contract")
			}
			mustPanic(t, func() { l.Sort() })
			mustPanic(t, func() { l.Sort(nil) })
			mustPanic(t, func() { l.Sort(compare.Ordered[int](), compare.Ordered[int]()) })
			l.Add(3, 1, 3)
			l.Prepend(5)
			l.Append(7)
			if !l.Insert(2, 8, 9) || !l.Set(0, 6) || !l.Swap(0, 6) {
				t.Fatal("valid mutation failed")
			}
			if !slices.Equal(l.Values(), []int{7, 3, 8, 9, 1, 3, 6}) {
				t.Fatal(l.Values())
			}
			if l.IndexOf(3) != 1 || l.LastIndexOf(3) != 5 || !l.Contains(3, 6) {
				t.Fatal("equality lookup")
			}
			if l.Contains(99) || l.IndexOf(99) != -1 || l.LastIndexOf(99) != -1 {
				t.Fatal("missing lookup")
			}
			if l.Set(l.Size(), 4) || l.Insert(-1, 4) || l.Remove(99) || l.Swap(-1, 0) {
				t.Fatal("invalid index accepted")
			}
			if _, ok := l.Get(-1); ok {
				t.Fatal("invalid get")
			}
			if !l.RemoveValue(3) || l.RemoveValue(99) {
				t.Fatal("remove value")
			}
			copy := l.Values()
			copy[0] = -100
			if value, _ := l.Get(0); value == -100 {
				t.Fatal("snapshot alias")
			}
			l.Sort(compare.Ordered[int]())
			if !slices.IsSorted(l.Values()) {
				t.Fatal("sort")
			}
			if _, _, ok := l.Find(func(i, v int) bool { return false }); ok {
				t.Fatal("find")
			}
			i, _, ok := l.Find(func(i, v int) bool { return v == 9 })
			if !ok || i != l.Size()-1 {
				t.Fatal("find result")
			}
			if !l.Any(func(i, v int) bool { return v == 9 }) || !l.All(func(i, v int) bool { return v > 0 }) {
				t.Fatal("enumeration")
			}
			before := l.Values()
			if l.FromJSON([]byte(`[1,"bad"]`)) == nil || !slices.Equal(before, l.Values()) {
				t.Fatal("failed decode mutated list")
			}
			data, err := json.Marshal(l)
			if err != nil {
				t.Fatal(err)
			}
			other := newList()
			if err := json.Unmarshal(data, other); err != nil {
				t.Fatal(err)
			}
			if !l.Equal(other) || l.HashCode() != other.HashCode() {
				t.Fatal("roundtrip")
			}
			l.Clear()
			if !l.Empty() || l.String() != "[]" {
				t.Fatal("clear")
			}
			data, _ = l.ToJSON()
			if string(data) != "[]" {
				t.Fatal(string(data))
			}
			if err := l.FromJSON([]byte(`null`)); err != nil {
				t.Fatal(err)
			}
			if !l.Empty() {
				t.Fatal("null decode")
			}
			mustPanic(t, func() { l.Each(nil) })
			mustPanic(t, func() { l.Find(nil) })
		})
	}
}

func TestRandomOperations(t *testing.T) {
	for name, newList := range factories() {
		t.Run(name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))
			l := newList()
			var model []int
			for range 3000 {
				i := rng.Intn(len(model) + 1)
				v := rng.Intn(100)
				switch rng.Intn(6) {
				case 0:
					l.Insert(i, v)
					model = slices.Insert(model, i, v)
				case 1:
					if i < len(model) {
						l.Remove(i)
						model = slices.Delete(model, i, i+1)
					}
				case 2:
					if i < len(model) {
						l.Set(i, v)
						model[i] = v
					}
				case 3:
					l.Add(v)
					model = append(model, v)
				case 4:
					l.Prepend(v)
					model = slices.Insert(model, 0, v)
				case 5:
					if len(model) > 70 {
						l.Clear()
						model = nil
					}
				}
				if !slices.Equal(l.Values(), model) || l.Size() != len(model) {
					t.Fatal("storage diverged", l.Values(), model)
				}
				for j, want := range model {
					got, ok := l.Get(j)
					if !ok || got != want {
						t.Fatal("indexed access diverged")
					}
				}
			}
		})
	}
}

type identity struct {
	base.Base
	Name string
}
type person struct {
	ID   int
	Name string
}

func (p person) Equal(other any) bool   { q, ok := other.(person); return ok && p.ID == q.ID }
func (p person) HashCode() uint64       { return base.HashCode(p.ID) }
func (p person) CompareTo(q person) int { return compare.Ordered[int]()(p.ID, q.ID) }

type incomplete struct{}

func (incomplete) Equal(any) bool { return true }

func TestObjectsAndConstructors(t *testing.T) {
	a, b := &identity{Name: "x"}, &identity{Name: "x"}
	l := arraylist.New[*identity]()
	l.Add(a)
	if !l.Contains(a) || l.Contains(b) || !base.Equals(a, a) {
		t.Fatal("embedded identity")
	}
	if !a.Equal(a.BaseIdentity()) || a.HashCode() != a.BaseIdentity().HashCode() {
		t.Fatal("identity hash contract")
	}
	var nilObject *identity
	l.Add(nilObject)
	if !l.Contains(nilObject) {
		t.Fatal("typed nil")
	}
	p := arraylist.New[person]()
	p.Add(person{2, "b"}, person{1, "first"}, person{1, "second"})
	if !p.Contains(person{1, "different"}) {
		t.Fatal("custom equality")
	}
	p.Sort(compare.FromComparable[person]())
	if p.Values()[0].Name != "first" {
		t.Fatal("stable sort")
	}
	q := doublylinkedlist.New[person]()
	q.Add(p.Values()...)
	if !p.Equal(q) || !q.Equal(p) || p.HashCode() != q.HashCode() {
		t.Fatal("cross-list equality")
	}
	unhashed := arraylist.New[incomplete]()
	unhashed.Add(incomplete{})
	mustPanic(t, func() { unhashed.HashCode() })
	bad := arraylist.New[any]()
	bad.Add(1, struct{}{})
	if bad.Size() != 2 {
		t.Fatal("storage must accept arbitrary values")
	}
	if err := bad.FromJSON([]byte(`[{"a":1}]`)); err != nil {
		t.Fatal(err)
	}
	mustPanic(t, func() { arraylist.NewWithComparator[int](nil) })
	order := compare.Ordered[int]().Reversed()
	x := arraylist.NewWithComparator(order)
	y := singlylinkedlist.NewWithComparator(order)
	z := doublylinkedlist.NewWithComparator(order)
	for _, current := range []lists.ExtendedList[int]{x, y, z} {
		current.Add(1, 3, 2)
		current.Sort()
		if !slices.Equal(current.Values(), []int{3, 2, 1}) {
			t.Fatal("saved comparator")
		}
	}
	if x.Map(func(i, v int) int { return v * 2 }).Size() != 3 {
		t.Fatal("map")
	}
	if y.Select(func(i, v int) bool { return v > 1 }).Size() != 2 {
		t.Fatal("select")
	}
	z.Map(func(i, v int) int { return -v }).Sort()
	var zero singlylinkedlist.List[int]
	zero.Add(1, 2)
	if zero.Size() != 2 {
		t.Fatal("zero value")
	}
}

func TestIterators(t *testing.T) {
	a, s, d := arraylist.New[int](), singlylinkedlist.New[int](), doublylinkedlist.New[int]()
	for _, l := range []lists.ExtendedList[int]{a, s, d} {
		l.Add(1, 2, 3)
	}
	for _, c := range []containers.IteratorWithIndex[int]{a.Iterator(), s.Iterator(), d.Iterator()} {
		mustPanic(t, func() { c.Value() })
		for i := range 3 {
			if !c.Next() || c.Index() != i || c.Value() != i+1 {
				t.Fatal("forward cursor")
			}
		}
		if c.Next() || c.Next() {
			t.Fatal("exhaustion")
		}
		mustPanic(t, func() { c.Index() })
		c.Begin()
		if !c.NextTo(func(i, v int) bool { return v == 2 }) {
			t.Fatal("next to")
		}
		if !c.First() || c.Value() != 1 {
			t.Fatal("first")
		}
		mustPanic(t, func() { c.NextTo(nil) })
	}
	for _, c := range []containers.ReverseIteratorWithIndex[int]{a.Iterator(), d.Iterator()} {
		c.End()
		for i := 2; i >= 0; i-- {
			if !c.Prev() || c.Value() != i+1 {
				t.Fatal("reverse cursor")
			}
		}
		if c.Prev() {
			t.Fatal("reverse exhaustion")
		}
		if !c.Last() || c.Value() != 3 {
			t.Fatal("last")
		}
		if !c.PrevTo(func(i, v int) bool { return v == 1 }) {
			t.Fatal("prev to")
		}
	}
	c := a.Iterator()
	a.Set(0, 5)
	mustPanic(t, func() { c.Begin() })
	mustPanic(t, func() { d.Each(func(i, v int) { d.Add(4) }) })
	mustPanic(t, func() { s.Any(func(i, v int) bool { s.Clear(); return true }) })
	for _, c := range []containers.ReverseIteratorWithIndex[int]{
		arraylist.New[int]().Iterator(), doublylinkedlist.New[int]().Iterator(),
	} {
		if c.First() || c.Last() || c.Next() || c.Prev() {
			t.Fatal("empty cursor")
		}
	}
}

func TestTransformAndInvalidationEdges(t *testing.T) {
	order := compare.Ordered[int]()
	a := arraylist.NewWithComparator(order)
	s := singlylinkedlist.NewWithComparator(order)
	d := doublylinkedlist.NewWithComparator(order)
	for _, l := range []lists.ExtendedList[int]{a, s, d} {
		l.Add(3, 1, 2)
	}
	inc := func(i, v int) int { return v + 1 }
	positive := func(i, v int) bool { return v > 1 }
	results := []lists.ExtendedList[int]{a.Map(inc), s.Map(inc), d.Map(inc)}
	for _, l := range results {
		l.Sort()
		if !slices.Equal(l.Values(), []int{2, 3, 4}) {
			t.Fatal("map/default order")
		}
	}
	for _, l := range []lists.ExtendedList[int]{a.Select(positive), s.Select(positive), d.Select(positive)} {
		l.Sort()
		if !slices.Equal(l.Values(), []int{2, 3}) {
			t.Fatal("select/default order")
		}
	}
	for _, run := range []func(){
		func() { a.Map(nil) }, func() { a.Select(nil) },
		func() { s.Map(nil) }, func() { s.Select(nil) },
		func() { d.Map(nil) }, func() { d.Select(nil) },
		func() { singlylinkedlist.NewWithComparator[int](nil) },
		func() { doublylinkedlist.NewWithComparator[int](nil) },
	} {
		mustPanic(t, run)
	}
	before := a.Values()
	mustPanic(t, func() { a.Sort(func(x, y int) int { panic("comparison failure") }) })
	if !slices.Equal(a.Values(), before) {
		t.Fatal("sort panic changed source")
	}
	for _, mutation := range []func(){
		func() { a.Add(4) }, func() { a.Remove(0) },
		func() { a.Set(0, 9) }, func() { a.Swap(0, 1) },
		func() { a.Sort() }, func() { a.FromJSON([]byte(`[1,2,3]`)) },
		func() { a.Clear() },
	} {
		a.FromJSON([]byte(`[1,2,3]`))
		cursor := a.Iterator()
		mutation()
		mustPanic(t, func() { cursor.Next() })
	}
	d.Clear()
	d.Add(1, 2, 3)
	c := d.Iterator()
	for c.Next() {
	}
	if !c.Prev() || c.Value() != 3 {
		t.Fatal("reverse from exhausted")
	}
	if c.PrevTo(func(i, v int) bool { return false }) {
		t.Fatal("reverse miss")
	}
	if !c.Next() || c.Value() != 1 {
		t.Fatal("forward from reverse exhaustion")
	}
	mustPanic(t, func() { c.PrevTo(nil) })
	c.Begin()
	mustPanic(t, func() { c.NextTo(func(i, v int) bool { d.Clear(); return true }) })
}

func FuzzListOperations(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 9, 0, 2})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 512 {
			operations = operations[:512]
		}
		for _, newList := range factories() {
			l := newList()
			var model []int
			for _, operation := range operations {
				index := int(operation) % (len(model) + 1)
				value := int(operation)
				switch operation % 5 {
				case 0:
					l.Insert(index, value)
					model = slices.Insert(model, index, value)
				case 1:
					if index < len(model) {
						l.Remove(index)
						model = slices.Delete(model, index, index+1)
					}
				case 2:
					if index < len(model) {
						l.Set(index, value)
						model[index] = value
					}
				case 3:
					l.Sort(compare.Ordered[int]())
					slices.Sort(model)
				case 4:
					l.Prepend(value)
					model = slices.Insert(model, 0, value)
				}
				if !slices.Equal(l.Values(), model) {
					t.Fatal("model mismatch")
				}
			}
		}
	})
}
