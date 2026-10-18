package lists_test

import (
	"iter"
	"math/rand"
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/lists"
	"github.com/Shenghui-Waa/go.repo/lists/doublylinkedlist"
	"github.com/Shenghui-Waa/go.repo/lists/singlylinkedlist"
)

type endpointList[T any] interface {
	lists.ExtendedList[T]
	lists.Endpoints[T]
}

func TestLinkedEndpoints(t *testing.T) {
	for _, l := range []endpointList[int]{
		new(singlylinkedlist.List[int]), new(doublylinkedlist.List[int]),
	} {
		empty := l.Values()
		for _, read := range []func() (int, bool){
			l.GetFirst, l.GetLast, l.RemoveFirst, l.RemoveLast,
		} {
			if value, ok := read(); ok || value != 0 {
				t.Fatal("empty result")
			}
		}
		if !slices.Equal(empty, l.Values()) {
			t.Fatal("empty mutation")
		}
		l.AddFirst(2)
		l.AddFirst(1)
		l.AddLast(3)
		if !slices.Equal(l.Values(), []int{1, 2, 3}) {
			t.Fatal("endpoint ordering")
		}
		if v, ok := l.GetFirst(); !ok || v != 1 {
			t.Fatal("first")
		}
		if v, ok := l.GetLast(); !ok || v != 3 {
			t.Fatal("last")
		}
		if v, ok := l.RemoveLast(); !ok || v != 3 {
			t.Fatal("tail deletion")
		}
		if v, ok := l.RemoveFirst(); !ok || v != 1 {
			t.Fatal("head deletion")
		}
		if v, ok := l.RemoveLast(); !ok || v != 2 || !l.Empty() {
			t.Fatal("singleton")
		}
		l.AddLast(0)
		if v, ok := l.RemoveFirst(); !ok || v != 0 {
			t.Fatal("zero member")
		}
		l.AddFirst(4)
		l.AddLast(5)
		if !slices.Equal(l.Values(), []int{4, 5}) {
			t.Fatal("reuse after empty")
		}
	}
}

func TestEndpointArbitraryValues(t *testing.T) {
	for _, l := range []endpointList[[]int]{
		singlylinkedlist.New[[]int](), doublylinkedlist.New[[]int](),
	} {
		l.AddFirst(nil)
		l.AddLast([]int{1, 2})
		if v, ok := l.RemoveFirst(); !ok || v != nil {
			t.Fatal("nil member")
		}
		if v, ok := l.RemoveLast(); !ok || !slices.Equal(v, []int{1, 2}) {
			t.Fatal("noncomparable member")
		}
	}
}

func TestEndpointIteratorInvalidation(t *testing.T) {
	for _, l := range []interface {
		endpointList[int]
		Seq() iter.Seq2[int, int]
	}{singlylinkedlist.New[int](), doublylinkedlist.New[int]()} {
		empty := l.Seq()
		l.RemoveFirst()
		l.RemoveLast()
		for range empty {
			t.Fatal("empty traversal")
		}
		for _, mutate := range []func(){
			func() { l.AddFirst(1) }, func() { l.AddLast(2) },
			func() { l.RemoveFirst() }, func() { l.RemoveLast() },
		} {
			l.Clear()
			l.Add(10, 20)
			seq := l.Seq()
			l.GetFirst()
			l.GetLast()
			count := 0
			for range seq {
				count++
			}
			if count != 2 {
				t.Fatal("read invalidated traversal")
			}
			mutate()
			mustPanic(t, func() {
				for range seq {
				}
			})
		}
	}
}

func TestRandomEndpoints(t *testing.T) {
	for _, l := range []endpointList[int]{
		singlylinkedlist.New[int](), doublylinkedlist.New[int](),
	} {
		rng := rand.New(rand.NewSource(18))
		var model []int
		for range 3000 {
			value := rng.Intn(100)
			switch rng.Intn(4) {
			case 0:
				l.AddFirst(value)
				model = slices.Insert(model, 0, value)
			case 1:
				l.AddLast(value)
				model = append(model, value)
			case 2:
				got, ok := l.RemoveFirst()
				if ok != (len(model) > 0) {
					t.Fatal("first status")
				}
				if ok {
					if got != model[0] {
						t.Fatal("first value")
					}
					model = model[1:]
				}
			case 3:
				got, ok := l.RemoveLast()
				if ok != (len(model) > 0) {
					t.Fatal("last status")
				}
				if ok {
					if got != model[len(model)-1] {
						t.Fatal("last value")
					}
					model = model[:len(model)-1]
				}
			}
			if !slices.Equal(l.Values(), model) {
				t.Fatal("model mismatch")
			}
		}
	}
}
