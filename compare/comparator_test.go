package compare

import (
	"cmp"
	"math"
	"slices"
	"testing"
)

func TestOrdered(t *testing.T) {
	order := Ordered[int]()
	values := []int{3, -1, 2, 2}
	slices.SortFunc(values, order)
	if !slices.Equal(values, []int{-1, 2, 2, 3}) {
		t.Fatal(values)
	}
	if order.Compare(2, 2) != 0 || order.Compare(2, 1) <= 0 {
		t.Fatal("incorrect Compare method")
	}
	type label string
	if Ordered[label]()("a", "b") >= 0 {
		t.Fatal("named ordered types must be supported")
	}
	floating := Ordered[float64]()
	if floating(math.NaN(), 0) >= 0 || floating(math.NaN(), math.NaN()) != 0 {
		t.Fatal("incorrect NaN ordering")
	}
	if floating(0, math.Copysign(0, -1)) != 0 {
		t.Fatal("signed zeros must compare equally")
	}
}

func TestReversed(t *testing.T) {
	minimum := -int(^uint(0)>>1) - 1
	order := Comparator[int](func(left, right int) int {
		if left < right {
			return minimum
		}
		return cmp.Compare(left, right)
	})
	if order.Reversed()(2, 1) >= 0 || order.Reversed()(1, 2) <= 0 {
		t.Fatal("reversal must not negate and overflow")
	}
	if order.Reversed().Reversed()(1, 2) != minimum {
		t.Fatal("double reversal must restore the comparator")
	}
}

func TestComparableComposition(t *testing.T) {
	values := []comparableTask{{3}, {1}, {2}}
	slices.SortFunc(values, FromComparable[comparableTask]().Reversed())
	if !slices.Equal(values, []comparableTask{{3}, {2}, {1}}) {
		t.Fatal(values)
	}
	secondaryCalls := 0
	secondary := Comparator[comparableTask](func(left, right comparableTask) int {
		secondaryCalls++
		return 0
	})
	order := FromComparable[comparableTask]().ThenComparing(secondary)
	order(comparableTask{1}, comparableTask{2})
	if secondaryCalls != 0 {
		t.Fatal("secondary comparator must short circuit")
	}
	order(comparableTask{1}, comparableTask{1})
	if secondaryCalls != 1 {
		t.Fatal("secondary comparator must break ties")
	}
}

func TestComparing(t *testing.T) {
	type task struct {
		priority int
		name     string
	}
	byPriority := ComparingOrdered(func(value task) int { return value.priority })
	byName := Comparing(func(value task) string { return value.name }, Ordered[string]())
	order := byPriority.ThenComparing(byName.Reversed())
	values := []task{{2, "a"}, {1, "a"}, {1, "b"}}
	slices.SortStableFunc(values, order)
	if !slices.Equal(values, []task{{1, "b"}, {1, "a"}, {2, "a"}}) {
		t.Fatal(values)
	}
	byNatural := Comparing(
		func(value task) comparableTask { return comparableTask{value.priority} },
		FromComparable[comparableTask](),
	)
	if byNatural(task{1, ""}, task{2, ""}) >= 0 {
		t.Fatal("keys must support Comparable adapters")
	}
}

func TestNilOrdering(t *testing.T) {
	a, b := 1, 2
	order := Comparator[*int](func(left, right *int) int {
		return cmp.Compare(*left, *right)
	})
	for _, nilFirst := range []bool{true, false} {
		wrapped := NilLast(order)
		want := []*int{&a, &b, nil, nil}
		if nilFirst {
			wrapped = NilFirst(order)
			want = []*int{nil, nil, &a, &b}
		}
		values := []*int{&b, nil, &a, nil}
		slices.SortFunc(values, wrapped)
		if !slices.Equal(values, want) || wrapped(nil, nil) != 0 {
			t.Fatal("incorrect nil ordering")
		}
	}
}

func TestNilFunctionsRejected(t *testing.T) {
	var empty Comparator[int]
	cases := map[string]func(){
		"reverse":   func() { empty.Reversed() },
		"primary":   func() { empty.ThenComparing(Ordered[int]()) },
		"secondary": func() { Ordered[int]().ThenComparing(empty) },
		"key":       func() { Comparing[int, int](nil, Ordered[int]()) },
		"key order": func() {
			Comparing(func(value int) int { return value }, empty)
		},
		"nil first": func() { NilFirst[int](nil) },
		"nil last":  func() { NilLast[int](nil) },
	}
	for name, run := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected immediate panic for nil function")
				}
			}()
			run()
		})
	}
}

func FuzzComparatorOrdering(f *testing.F) {
	f.Add(1, 2, 3)
	f.Add(0, 0, -1)
	f.Fuzz(func(t *testing.T, a, b, c int) {
		order := Ordered[int]().Reversed()
		if order(a, b) != -order(b, a) || order(a, a) != 0 {
			t.Fatal("ordering must be antisymmetric and reflexive")
		}
		if order(a, b) < 0 && order(b, c) < 0 && order(a, c) >= 0 {
			t.Fatal("ordering must be transitive")
		}
	})
}
