package lists_test

import (
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/lists"
	"github.com/Shenghui-Waa/go.repo/lists/arraylist"
	"github.com/Shenghui-Waa/go.repo/lists/doublylinkedlist"
	"github.com/Shenghui-Waa/go.repo/lists/singlylinkedlist"
)

func TestArbitraryListStorage(t *testing.T) {
	for _, l := range []lists.ExtendedList[[]int]{
		arraylist.New[[]int](), singlylinkedlist.New[[]int](),
		doublylinkedlist.New[[]int](),
	} {
		l.Add([]int{3}, nil)
		l.Insert(1, []int{2})
		l.Set(0, []int{1})
		if value, ok := l.Get(0); !ok || !slices.Equal(value, []int{1}) {
			t.Fatal("slice storage")
		}
		l.Sort(slices.Compare[[]int, int])
		if l.Size() != 3 || !l.Remove(0) {
			t.Fatal("positional operations")
		}
		if err := l.FromJSON([]byte(`[[4],[5]]`)); err != nil {
			t.Fatal(err)
		}
		if !lists.ContainsFunc[[]int](l, []int{5}, slices.Equal[[]int, int]) {
			t.Fatal("explicit equality lookup")
		}
		if lists.IndexOfFunc[[]int](l, []int{4}, slices.Equal[[]int, int]) != 0 {
			t.Fatal("explicit index")
		}
		mustPanic(t, func() { l.Contains([]int{4}) })
		mustPanic(t, func() { l.HashCode() })
		mustPanic(t, func() { l.Equal(l) })
	}
}

func TestExplicitEqualityAndHash(t *testing.T) {
	a, b := arraylist.New[[]int](), singlylinkedlist.New[[]int]()
	a.Add([]int{1, 2}, []int{3})
	b.Add([]int{1, 2}, []int{3})
	equal := slices.Equal[[]int, int]
	hash := func(values []int) uint64 {
		result := uint64(1)
		for _, value := range values {
			result = 31*result + uint64(value)
		}
		return result
	}
	if !lists.EqualFunc[[]int](a, b, equal) {
		t.Fatal("explicit list equality")
	}
	if lists.HashFunc[[]int](a, hash) != lists.HashFunc[[]int](b, hash) {
		t.Fatal("explicit hash contract")
	}
	b.Set(1, []int{4})
	if lists.EqualFunc[[]int](a, b, equal) {
		t.Fatal("unequal elements")
	}
	b.Clear()
	if lists.EqualFunc[[]int](a, b, equal) {
		t.Fatal("unequal lengths")
	}
	if lists.HashFunc[[]int](b, hash) != 1 {
		t.Fatal("empty hash")
	}
	for _, run := range []func(){
		func() { lists.EqualFunc[[]int](b, b, nil) },
		func() { lists.HashFunc[[]int](b, nil) },
		func() { lists.IndexOfFunc[[]int](b, nil, nil) },
	} {
		mustPanic(t, run)
	}
}
