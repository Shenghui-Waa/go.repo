package compare

import "time"

// Time compares instants using time.Time.Compare, including its monotonic clock
// semantics. Location names do not determine ordering. This is not Go's ==
// comparison of time.Time fields and does not adapt time.Time to Comparable.
func Time() Comparator[time.Time] {
	return time.Time.Compare
}
