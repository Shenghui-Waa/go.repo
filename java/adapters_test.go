package java

import (
	"errors"
	"testing"

	"github.com/Shenghui-Waa/go.repo/arraylist"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/hashmap"
	"github.com/Shenghui-Waa/go.repo/iterable"
	"github.com/Shenghui-Waa/go.repo/linkedlist"
)

func TestQueueAdapter(t *testing.T) {
	queue := Queue[int]{Core: linkedlist.NewComparable[int]()}
	if _, err := queue.Element(); !errors.Is(err, iterable.ErrNoSuchElement) {
		t.Fatalf("err = %v", err)
	}
	queue.Add(3)
	if value, err := queue.Remove(); err != nil || value != 3 {
		t.Fatalf("remove = %v,%v", value, err)
	}
}

func TestListIteratorAdapter(t *testing.T) {
	adapter := List[int]{Core: arraylist.NewComparable(1, 2)}
	iterator := adapter.Iterator()
	value, err := iterator.Next()
	if err != nil || value != 1 {
		t.Fatalf("next = %v,%v", value, err)
	}
	if err := iterator.Remove(); err != nil {
		t.Fatal(err)
	}
	if adapter.Size() != 1 {
		t.Fatalf("size = %d", adapter.Size())
	}
}

func TestMapViewsWriteThrough(t *testing.T) {
	core := hashmap.New[string, int]()
	core.Put("a", 1)
	mapping := NewMap[string, int](core, compare.NativeEqual[int])
	if !mapping.KeySet().Remove("a") || core.ContainsKey("a") {
		t.Fatal("key set did not write through")
	}
	core.Put("b", 2)
	for entry := range mapping.EntrySet().All() {
		if _, err := entry.SetValue(3); err != nil {
			t.Fatal(err)
		}
	}
	if value, _ := core.Get("b"); value != 3 {
		t.Fatalf("value = %d", value)
	}
}
