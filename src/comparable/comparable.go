package comparable

// Comparable 比较接口
type Comparable[T any] interface {
	Compare(other T) int // 返回负数表示当前元素优先级更高
}
