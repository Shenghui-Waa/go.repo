// Package queues defines FIFO and priority-based removal contracts.
package queues

import (
	"fmt"

	"github.com/Shenghui-Waa/go.repo/containers"
)

// Queue permits duplicate and arbitrary values. Removal order is FIFO except
// for priority queues. Implementations are not safe for concurrent use.
type Queue[T any] interface {
	containers.Container[T]
	Enqueue(T)
	Dequeue() (T, bool)
	Peek() (T, bool)
}

type ExtendedQueue[T any] interface {
	Queue[T]
	containers.Clearer
	containers.EnumerableWithIndex[T]
	containers.IndexedSequence[T]
	containers.JSONCodec
	fmt.Stringer
}
