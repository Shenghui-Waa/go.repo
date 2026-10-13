package priorityqueue

import (
	"container/heap"

	"github.com/Shenghui-Waa/go.repo/comparable"
)

type PriorityQueue[T comparable.Comparable[T]] []T

// New 创建优先队列
func New[T comparable.Comparable[T]]() *PriorityQueue[T] {
	pq := make(PriorityQueue[T], 0)
	heap.Init(&pq)
	return &pq
}

func (pq *PriorityQueue[T]) Len() int {
	return len(*pq)
}

// Less 调用元素的 Compare 方法
func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return (*pq)[i].Compare((*pq)[j]) < 0
}

func (pq *PriorityQueue[T]) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
}

func (pq *PriorityQueue[T]) Push(x any) {
	*pq = append(*pq, x.(T))
}

func (pq *PriorityQueue[T]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

// Offer / Poll 便捷方法
func (pq *PriorityQueue[T]) Offer(item T) {
	heap.Push(pq, item)
}

func (pq *PriorityQueue[T]) Poll() T {
	return heap.Pop(pq).(T)
}
