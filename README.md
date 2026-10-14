# go.repo

`go.repo` is a Go 1.25 data-structure library whose contracts are derived from
the Java 25 Collections Framework while retaining idiomatic Go APIs.

## Design

The library has two layers:

- Core packages use generics, `iter.Seq`, `iter.Seq2`, `(value, ok)` results,
  and function-based comparison strategies.
- Package `java` exposes Java-shaped adapters such as iterator, list, queue,
  map, and live map views.

Containers are not concurrency-safe. Structural mutation while iterating is
detected on ordered and indexed implementations and results in
`iterable.ErrConcurrentModification`. A comparator result of zero means that
two sorted-map keys or sorted-set elements are the same, even when another Go
equality rule would distinguish them.

## Package map

| Java concept | Go contract or implementation |
| --- | --- |
| `Comparable`, `Comparator` | `compare.Comparable`, `compare.Comparator` |
| `Iterable`, `Iterator` | `iterable.Iterable`, `iterable.Iterator` |
| `Collection`, `List` | `collections.Collection`, `collections.List` |
| `Queue`, `Deque` | `collections.Queue`, `collections.Deque` |
| `Set`, `NavigableSet` | `collections.Set`, `collections.NavigableSet` |
| `Map`, `NavigableMap` | `collections.Map`, `collections.NavigableMap` |
| `ArrayList` | `arraylist.ArrayList` |
| `LinkedList` | `linkedlist.LinkedList` |
| `PriorityQueue` | `priorityqueue.PriorityQueue` |
| `HashMap`, `HashSet` | `hashmap.HashMap`, `hashset.HashSet` |
| `TreeMap`, `TreeSet` | `treemap.TreeMap`, `treeset.TreeSet` |

`TreeMap` uses a red-black tree. `TreeSet` delegates storage and ordering to a
`TreeMap[T, struct{}]`.

## Examples

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

For values that implement the Java-shaped `CompareTo(T) int` contract, the
empty-queue construction style matches Java's natural-order constructor:

```go
queue := priorityqueue.New[Task]()
```

Use `priorityqueue.NewWithComparator` when ordering is external to the value.

## Empty-result and iterator rules

- Core queue and navigation methods return `(zero, false)` for an empty result.
- Java queue adapters return `iterable.ErrNoSuchElement` for `Element` and
  `Remove` on an empty queue.
- Core `All` and `Backward` sequences are read-only.
- `ArrayList.Positions` and `LinkedList.Positions` provide mutable,
  bidirectional list positions with fail-fast structural-modification checks.
- `PriorityQueue.All` exposes heap storage order, not sorted order.

## Migration from the original priority queue

The original API required every element to implement a custom `Compare` method
and exposed its backing slice. The new API deliberately breaks compatibility:

```go
// Before:
queue := priorityqueue.New[Item](func(left, right Item) int {
	return cmp.Compare(left.Priority, right.Priority)
})
value := queue.Poll()

// After:
queue := priorityqueue.NewWithComparator(func(left, right Item) int {
	return cmp.Compare(left.Priority, right.Priority)
})
value, ok := queue.Poll()
```

The backing heap is private, and custom ordering is supplied at construction.

## Verification

```powershell
$env:GOCACHE = '..\.temp\go-build'
go test ./...
go vet ./...
```
