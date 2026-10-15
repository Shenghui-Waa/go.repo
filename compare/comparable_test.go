package compare

import (
	"cmp"
	"testing"

	"github.com/emirpasic/gods/utils"
)

type comparableTask struct {
	priority int
}

func (task comparableTask) CompareTo(other comparableTask) int {
	return cmp.Compare(task.priority, other.priority)
}

func TestFromComparableAdaptsToGoDSComparator(t *testing.T) {
	var comparator utils.Comparator = FromComparable[comparableTask]()

	if comparator(
		comparableTask{priority: 1},
		comparableTask{priority: 2},
	) >= 0 {
		t.Fatal("comparable adapter returned the wrong order")
	}
}
