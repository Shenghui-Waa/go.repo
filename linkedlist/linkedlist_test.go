package linkedlist

import (
	"slices"
	"testing"
)

func TestLinkedListListAndDeque(t *testing.T) {
	list := NewComparable(2, 3)
	list.OfferFirst(1)
	list.OfferLast(4)
	if value, ok := list.PollFirst(); !ok || value != 1 {
		t.Fatalf("first = %v", value)
	}
	if value, ok := list.PollLast(); !ok || value != 4 {
		t.Fatalf("last = %v", value)
	}
	list.Insert(1, 9)
	if got := slices.Collect(list.All()); !slices.Equal(got, []int{2, 9, 3}) {
		t.Fatalf("all = %v", got)
	}
	position := list.Positions(1)
	if err := position.Add(8); err != nil {
		t.Fatal(err)
	}
	if got := slices.Collect(list.All()); !slices.Equal(got, []int{2, 8, 9, 3}) {
		t.Fatalf("after add = %v", got)
	}
}
