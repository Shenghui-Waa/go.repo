# go.repo

`go.repo` 是一个基于 Go 1.25 实现的数据结构库。其接口契约派生自
Java 25 Collections Framework，同时保留符合 Go 语言习惯的 API。

## 设计

本库分为两个层次：

- 核心包使用泛型、`iter.Seq`、`iter.Seq2`、`(value, ok)` 返回值以及
  基于函数的比较策略。
- `java` 包提供接近 Java 使用方式的适配器，包括迭代器、列表、队列、
  Map 以及可回写的 Map 视图。

容器不保证并发安全。有序容器和索引容器会检测迭代期间发生的结构性修改，
并触发 `iterable.ErrConcurrentModification`。对于有序 Map 和有序 Set，
比较器返回零表示两个键或元素相同，即使其他 Go 相等性规则可能将它们视为不同值。

## 包与类型映射

| Java 概念 | Go 接口或实现 |
| --- | --- |
| `Comparable`、`Comparator` | `compare.Comparable`、`compare.Comparator` |
| `Iterable`、`Iterator` | `iterable.Iterable`、`iterable.Iterator` |
| `Collection`、`List` | `collections.Collection`、`collections.List` |
| `Queue`、`Deque` | `collections.Queue`、`collections.Deque` |
| `Set`、`NavigableSet` | `collections.Set`、`collections.NavigableSet` |
| `Map`、`NavigableMap` | `collections.Map`、`collections.NavigableMap` |
| `ArrayList` | `arraylist.ArrayList` |
| `LinkedList` | `linkedlist.LinkedList` |
| `PriorityQueue` | `priorityqueue.PriorityQueue` |
| `HashMap`、`HashSet` | `hashmap.HashMap`、`hashset.HashSet` |
| `TreeMap`、`TreeSet` | `treemap.TreeMap`、`treeset.TreeSet` |

`TreeMap` 使用红黑树实现。`TreeSet` 将元素存储与排序委托给
`TreeMap[T, struct{}]`。

## 使用示例

```go
queue := priorityqueue.NewOrdered(4, 1, 3)
value, ok := queue.Poll() // 1, true

list := arraylist.NewComparable("a", "b")
for value := range list.All() {
	fmt.Println(value)
}

mapping := treemap.New[int, string](compare.Natural[int])
mapping.Put(2, "two")
mapping.Put(1, "one")
for key, value := range mapping.All() {
	fmt.Println(key, value)
}
```

对于实现了 Java 风格 `CompareTo(T) int` 接口的类型，可以使用与 Java
自然排序构造器相近的空队列创建方式：

```go
queue := priorityqueue.New[Task]()
```

当排序规则由类型外部提供时，请使用 `priorityqueue.NewWithComparator`。

## 空结果与迭代规则

- 核心队列和导航方法在没有结果时返回 `(zero, false)`。
- Java 队列适配器在空队列上调用 `Element` 或 `Remove` 时，返回
  `iterable.ErrNoSuchElement`。
- 核心层的 `All` 和 `Backward` 序列均为只读迭代序列。
- `ArrayList.Positions` 和 `LinkedList.Positions` 提供可修改、可双向移动的
  列表位置，并带有 fail-fast 结构修改检测。
- `PriorityQueue.All` 暴露的是堆的内部存储顺序，而不是完整排序后的顺序。

## 从旧版优先队列迁移

旧版 API 要求所有元素实现自定义 `Compare` 方法，并且直接暴露底层切片。
新版 API 有意进行了不兼容调整：

```go
// 旧版：
queue := priorityqueue.New[Item](func(left, right Item) int {
	return cmp.Compare(left.Priority, right.Priority)
})
value := queue.Poll()

// 新版：
queue := priorityqueue.NewWithComparator(func(left, right Item) int {
	return cmp.Compare(left.Priority, right.Priority)
})
value, ok := queue.Poll()
```

新版隐藏了底层堆存储，自定义排序规则需要在构造队列时传入。

如果元素自身实现了 `compare.Comparable[T]`，可以直接创建空优先队列：

```go
type Task struct {
	Priority int
}

func (task Task) CompareTo(other Task) int {
	return cmp.Compare(task.Priority, other.Priority)
}

queue := priorityqueue.New[Task]()
```

## 验证

```powershell
$env:GOCACHE = '..\.temp\go-build'
go test ./...
go vet ./...
```
