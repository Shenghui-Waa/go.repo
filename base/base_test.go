package base

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"unsafe"
)

type record struct{ id int }

func (r record) Equal(other any) bool {
	value, ok := other.(record)
	return ok && r.id == value.id
}

func (r record) HashCode() uint64 { return HashCode(r.id) }

type equalityOnly struct{}

func (e equalityOnly) Equal(other any) bool { return true }

func TestBaseIdentity(t *testing.T) {
	var _ CompareInterface = (*Base)(nil)
	var _ StringerInterface = (*Base)(nil)
	var _ fmt.Stringer = (*Base)(nil)
	left, right := New(), New()
	if !Equals(left, left) || Equals(left, right) || Equals(left, *left) {
		t.Fatal("Base must use pointer identity")
	}
	if HashCode(left) != left.HashCode() {
		t.Fatal("HashCode must delegate to Base")
	}
	validPrefix := strings.HasPrefix(left.String(), "*base.Base@")
	if left.ToString() != fmt.Sprint(left) || !validPrefix {
		t.Fatal("String and ToString must agree")
	}
	var empty *Base
	if !empty.Equal(nil) || empty.HashCode() != 0 || empty.String() != "<nil>" {
		t.Fatal("nil Base methods must be safe")
	}
}

func TestNil(t *testing.T) {
	values := []any{
		nil, (*int)(nil), (*Base)(nil), []int(nil), map[string]int(nil),
		(chan int)(nil), (func())(nil), unsafe.Pointer(nil),
	}
	for _, value := range values {
		if !IsNil(value) || NonNil(value) || !Equals(nil, value) {
			t.Fatalf("expected nil semantics for %T", value)
		}
		if HashCode(value) != 0 || !DeepEquals(value, nil) {
			t.Fatalf("expected nil hash and deep equality for %T", value)
		}
	}
	for _, value := range []any{0, false, "", []int{}, map[int]int{}, New()} {
		if IsNil(value) || !NonNil(value) || Equals(value, nil) {
			t.Fatalf("unexpected nil semantics for %T", value)
		}
	}
}

func TestEquals(t *testing.T) {
	cases := []struct {
		left, right any
		want        bool
	}{
		{1, 1, true}, {1, int64(1), false}, {1, 2, false},
		{record{7}, record{7}, true}, {record{7}, record{8}, false},
		{[]int{1}, []int{1}, false},
		{struct{ X any }{[]int{1}}, struct{ X any }{[]int{1}}, false},
		{struct{ X any }{1}, struct{ X any }{[]int{1}}, false},
	}
	for _, tc := range cases {
		if Equals(tc.left, tc.right) != tc.want {
			t.Fatalf("Equals(%v, %v): want %v", tc.left, tc.right, tc.want)
		}
	}
}

func TestDeepEquals(t *testing.T) {
	left := map[string][]int{"a": {1, 2}}
	right := map[string][]int{"a": {1, 2}}
	if !DeepEquals(left, right) || DeepEquals([]int(nil), []int{}) {
		t.Fatal("incorrect collection equality")
	}
	type node struct{ next *node }
	a, b := &node{}, &node{}
	a.next, b.next = a, b
	if !DeepEquals(a, b) {
		t.Fatal("equivalent cycles must compare equal")
	}
}

func TestHashContract(t *testing.T) {
	negativeZero := math.Copysign(0, -1)
	value := 7
	channel := make(chan int)
	pairs := [][2]any{
		{true, true}, {false, false}, {uint64(9), uint64(9)},
		{float32(0), float32(negativeZero)},
		{&value, &value}, {channel, channel},
		{unsafe.Pointer(&value), unsafe.Pointer(&value)},
		{struct{ X any }{}, struct{ X any }{}},
		{struct{ _ int }{}, struct{ _ int }{}},
		{0.0, negativeZero},
		{complex(0.0, 0.0), complex(negativeZero, negativeZero)},
		{record{4}, record{4}},
		{[2]string{"a", "bc"}, [2]string{"a", "bc"}},
		{struct{ private any }{3}, struct{ private any }{3}},
	}
	for _, pair := range pairs {
		if !Equals(pair[0], pair[1]) || HashCode(pair[0]) != HashCode(pair[1]) {
			t.Fatalf("equal values must hash equally: %v", pair)
		}
	}
	if Hash() != 1 || Hash(nil) != 31 || Hash(7) != 31+HashCode(7) {
		t.Fatal("incorrect combination semantics")
	}
	if Hash(1, "a") != 31*(31+HashCode(1))+HashCode("a") {
		t.Fatal("incorrect ordered hash combination")
	}
	if HashCode([2]string{"a", "bc"}) == HashCode([2]string{"ab", "c"}) {
		t.Fatal("strings must be length framed")
	}
}

func TestHashUnsupportedValues(t *testing.T) {
	values := []any{
		[]int{1}, map[int]int{}, func() {}, equalityOnly{},
		struct{ X any }{[]int{1}},
	}
	for _, value := range values {
		t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("unsupported hash input must panic")
				}
			}()
			HashCode(value)
		})
	}
}

func FuzzHashEquality(f *testing.F) {
	f.Add("abc", int64(42), float64(0))
	f.Add("", int64(-1), math.Copysign(0, -1))
	f.Fuzz(func(t *testing.T, text string, number int64, floating float64) {
		left := struct {
			text     string
			number   int64
			floating float64
		}{text, number, floating}
		right := left
		if Equals(left, right) && HashCode(left) != HashCode(right) {
			t.Fatal("equal values must hash equally")
		}
	})
}
