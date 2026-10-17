package compare_test

import (
	"slices"
	"testing"
	"time"

	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/queues/priorityqueue"
)

func TestTimeComparator(t *testing.T) {
	order := compare.Time()
	instant := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	earlier := instant.Add(-time.Second)
	later := instant.Add(time.Second)
	zone := instant.In(time.FixedZone("offset", 8*60*60))
	if order(instant, zone) != 0 || order(earlier, instant) >= 0 {
		t.Fatal("instant/location comparison")
	}
	if order(later, instant) <= 0 || order(time.Time{}, time.Time{}) != 0 {
		t.Fatal("later/zero comparison")
	}
	monotonic := time.Now()
	if order(monotonic, monotonic.Add(time.Second)) >= 0 {
		t.Fatal("monotonic comparison")
	}
	if order(monotonic, monotonic.Round(0)) != 0 {
		t.Fatal("mixed wall and monotonic comparison")
	}
	values := []time.Time{earlier, later, instant}
	slices.SortFunc(values, order.Reversed())
	if !values[0].Equal(later) || !values[2].Equal(earlier) {
		t.Fatal("reversed time sorting")
	}
	queue := priorityqueue.NewWithComparator(order)
	queue.Enqueue(later)
	queue.Enqueue(earlier)
	if got, ok := queue.Dequeue(); !ok || !got.Equal(earlier) {
		t.Fatal("priority queue integration")
	}
}
