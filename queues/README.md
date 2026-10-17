# Queues

Four implementations share queues.Queue[T], the local container contracts and
typed comparators. Original gods/ sources are unchanged. No external GoDS module
is required, and no GoDS implementation is used by queues.

| Package | Storage | Removal order | Constructor |
| --- | --- | --- | --- |
| arrayqueue | Growing ring array | FIFO | New[T]() |
| linkedlistqueue | Nodes with head/tail | FIFO | New[T]() |
| circularbuffer | Fixed ring | FIFO, overwrites oldest when full | New[T]() / NewWithCapacity[T](n) |
| priorityqueue | Local binaryheap | Minimum in comparator order | New[T]() / NewWithComparator(order) |

Ordinary queues support arbitrary types, duplicates and nil values. They do not
perform equality/hash operations, so Base, Equal and HashCode are not required.
Priority New[T]() requires CompareTo(T) int; use NewWithComparator for built-in
types or custom orders. Reverse the comparator for maximum-first removal.
Comparators must be deterministic, transitive, pure and non-panicking. Equal
priority does not promise stable FIFO tie ordering.

```go
fifo := arrayqueue.New[[]int]()
fifo.Enqueue([]int{1, 2})
value, ok := fifo.Dequeue()

buffer := circularbuffer.NewWithCapacity[string](32)
priority := priorityqueue.NewWithComparator(compare.Ordered[int]())
```

## Capacity and complexity

Circular New defaults to capacity 16; explicit capacity must be positive. Full
Enqueue discards the oldest value, matching GoDS. Capacity 1 is supported.
Capacity/Full describe a fixed bound: unbounded queues report 0/false, not their
allocated storage size. Circular Clear preserves capacity. Array Clear releases
its ring allocation. Removed ring slots are cleared to release references.

Array enqueue is amortized O(1), dequeue O(1). Linked enqueue/dequeue are O(1).
Circular operations are O(1). Priority enqueue/dequeue are O(log n), peek O(1).
FIFO array and node implementations are local: using the existing lists would
unnecessarily require their element equality contracts.

## Traversal and transformations

Values returns a shallow independent slice in FIFO order, except priorityqueue
which exposes heap layout, not globally sorted order. Use repeated Dequeue for
priority order. Seq and Iterator capture O(n) shallow snapshots; iteration steps
are O(1). All four provide bidirectional snapshot iterators, including the linked
queue. Index is a traversal position, not a persistent identifier.

Enqueue, successful Dequeue, nonempty Clear and JSON replacement invalidate
existing traversals. Peek, empty Dequeue and empty Clear do not. Invalid cursor
Value/Index access panics. Begin/End do not restore invalidated cursors. Callback
mutation is detected even on early termination. A failed priority operation may
also invalidate traversal because its comparator can panic during mutation;
regular mutation does not promise rollback on comparator failure.

Each/Any/All/Find/Map/Select reject nil callbacks. Any/All/Find short-circuit;
empty Any is false and empty All true. Find misses return (-1, zero, false).
Map and Select preserve queue kind, fixed capacity and comparator. Priority
results rebuild heap order. There is no automatic object equality or hash API.

## Serialization and lifecycle

JSON is a value array, empty as []; null clears. Decoding builds a temporary
queue and commits on success, retaining the original on malformed input or a
comparator panic. Circular input longer than capacity retains its newest values.
Priority input is reheapified; JSON does not preserve stable priority ties.
Interface decoding follows encoding/json and cannot infer custom object types.

Array, linked and circular zero values are usable; circular zero values acquire
the default capacity. Priority queues require a constructor. Queues are not
concurrency-safe. Do not copy populated queues or mutate stored priority fields.
Snapshots share referenced objects, not ownership of their contents.

Tests exercise FIFO and priority models, wraparound, array growth, bounded
overwrite, duplicate/nil/non-comparable values, transformations, invalidation
and failed-decode atomicity. Randomized tests cover all four storage variants.
