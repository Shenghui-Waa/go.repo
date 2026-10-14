package priorityqueue

import "testing"

func BenchmarkPriorityQueueOfferPoll(b *testing.B) {
	for iteration := 0; iteration < b.N; iteration++ {
		queue := NewOrdered[int]()
		for value := 1000; value >= 0; value-- {
			queue.Offer(value)
		}
		for !queue.IsEmpty() {
			queue.Poll()
		}
	}
}
