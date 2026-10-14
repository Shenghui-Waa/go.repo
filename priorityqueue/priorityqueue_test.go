package priorityqueue

import (
	"cmp"
	"math/rand"
	"sort"
	"testing"
)

type task struct {
	priority int
}

func (current task) CompareTo(other task) int {
	return cmp.Compare(current.priority, other.priority)
}

func TestPriorityQueueComparableConstructor(t *testing.T) {
	queue := New[task]()
	queue.Offer(task{priority: 3})
	queue.Offer(task{priority: 1})

	value, ok := queue.Poll()
	if !ok || value.priority != 1 {
		t.Fatalf("poll = %+v, %v", value, ok)
	}
}

func TestPriorityQueueOrderedPoll(t *testing.T) {
	random := rand.New(rand.NewSource(42))
	queue := NewOrdered[int]()
	want := make([]int, 500)
	for index := range want {
		want[index] = random.Intn(1000)
		queue.Offer(want[index])
	}
	sort.Ints(want)
	for _, expected := range want {
		value, ok := queue.Poll()
		if !ok || value != expected {
			t.Fatalf("poll = %v, want %v", value, expected)
		}
	}
	if _, ok := queue.Poll(); ok {
		t.Fatal("empty poll succeeded")
	}
}
