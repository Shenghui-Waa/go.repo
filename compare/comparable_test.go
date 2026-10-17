package compare

import (
	"cmp"
	"testing"
)

type comparableTask struct {
	priority int
}

func (task comparableTask) CompareTo(other comparableTask) int {
	return cmp.Compare(task.priority, other.priority)
}

func TestFromComparable(t *testing.T) {
	var comparator Comparator[comparableTask] = FromComparable[comparableTask]()

	if comparator(
		comparableTask{priority: 1},
		comparableTask{priority: 2},
	) >= 0 {
		t.Fatal("comparable adapter returned the wrong order")
	}
	if comparator(comparableTask{2}, comparableTask{2}) != 0 {
		t.Fatal("equal priorities must compare equally")
	}
	if comparator(comparableTask{3}, comparableTask{2}) <= 0 {
		t.Fatal("higher priority must compare greater")
	}
}
