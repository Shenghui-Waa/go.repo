package arraylist

import (
	"errors"
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/iterable"
)

func TestArrayListContract(t *testing.T) {
	list := NewComparable(2, 3)
	list.AddFirst(1)
	list.AddLast(4)
	if !list.Insert(2, 9) {
		t.Fatal("insert failed")
	}
	if got := slices.Collect(list.All()); !slices.Equal(got, []int{1, 2, 9, 3, 4}) {
		t.Fatalf("all = %v", got)
	}
	old, ok := list.Set(2, 8)
	if !ok || old != 9 {
		t.Fatalf("set = %v, %v", old, ok)
	}
	removed, ok := list.RemoveAt(1)
	if !ok || removed != 2 {
		t.Fatalf("remove = %v, %v", removed, ok)
	}
	if got := slices.Collect(list.Backward()); !slices.Equal(got, []int{4, 3, 8, 1}) {
		t.Fatalf("backward = %v", got)
	}
}

func TestArrayListPositionAndFailFast(t *testing.T) {
	list := NewComparable(1, 2, 3)
	position := list.Positions(1)
	value, err := position.Next()
	if err != nil || value != 2 {
		t.Fatalf("next = %v, %v", value, err)
	}
	if err := position.Set(20); err != nil {
		t.Fatal(err)
	}
	if err := position.Remove(); err != nil {
		t.Fatal(err)
	}
	if err := position.Add(4); err != nil {
		t.Fatal(err)
	}
	list.Add(5)
	if _, err := position.Next(); !errors.Is(err, iterable.ErrConcurrentModification) {
		t.Fatalf("err = %v", err)
	}
}

func FuzzArrayListOperations(f *testing.F) {
	f.Add([]byte{1, 2, 3, 2})
	f.Fuzz(func(t *testing.T, input []byte) {
		list := NewComparable[int]()
		model := make([]int, 0)
		for _, value := range input {
			if value%3 == 0 && len(model) > 0 {
				index := int(value) % len(model)
				got, ok := list.RemoveAt(index)
				if !ok || got != model[index] {
					t.Fatalf("remove mismatch")
				}
				model = append(model[:index], model[index+1:]...)
			} else {
				list.Add(int(value))
				model = append(model, int(value))
			}
		}
		if got := slices.Collect(list.All()); !slices.Equal(got, model) {
			t.Fatalf("got %v want %v", got, model)
		}
	})
}
