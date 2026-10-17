# Trees

This package owns its search-tree algorithms locally. The AVL, red-black and
B-tree kernels are adapted from the reference GoDS source, with generic keys
relaxed to any and comparators redirected to this project's compare package.
They are not runtime wrappers around the external GoDS module. Upstream
attribution and license terms are retained in LICENSE.GoDS. The binary heap is
a local slice-backed implementation with sift-up, sift-down and bulk heapify.

The original gods/ checkout is unchanged. No external GoDS module is required;
the adapted source's license notices must still be retained.
The project's maps tree indexes now use this local red-black tree implementation.

## Construction

```go
// Built-in types use explicit project comparators.
avl := avltree.NewWithComparator[int, string](compare.Ordered[int]())
rb := redblacktree.NewWithComparator[int, string](compare.Ordered[int]())
b := btree.NewWithComparator[int, string](compare.Ordered[int]())
customOrder := btree.NewWithOrderAndComparator[int, string](5, compare.Ordered[int]())
heap := binaryheap.NewWithComparator(compare.Ordered[int]())
maxHeap := binaryheap.NewWithComparator(compare.Ordered[int]().Reversed())
```

All public New constructors take no arguments. Search-tree New[K,V]() requires
K to implement CompareTo(K) int; heap New[T]() requires CompareTo(T) int.
NewWithComparator accepts custom comparators or compare.FromComparable.
B-tree New defaults to order 32 (maximum children); NewWithOrder and
NewWithOrderAndComparator support orders >= 3. Nil comparators are rejected.
Use constructors: zero values do not install an implicit order.

## Semantics

- Search trees preserve sorted key order. Comparator zero means the same key;
  updates retain the original key object and replace its value. Keys can be
  non-comparable Go structs/slices when the comparator supports them.
- Equality/hash methods are not required because these structures do not use
  hash lookup or object equality. Base embedding remains optional. Comparator
  equivalence should be consistent with object equality when both are used.
- Values can be arbitrary types. Heap duplicates are retained; equal priorities
  do not promise stable extraction order. Heap Values/Seq use heap layout, not
  sorted order. Repeated Pop returns priority order.
- Comparators must be deterministic, transitive, pure and non-panicking. Nil
  handling belongs to the comparator. Mutating key/priority fields in place or
  mutating the container from a comparison callback is unsupported.
- Public nodes and comparator fields are not exposed. GetNode/Root and mutable
  node access from GoDS are deliberately not reproduced.
- Put/Get/Remove are logarithmic for AVL/red-black trees. B-trees search nodes
  using binary search; splitting/merging includes O(order) node work. Remove
  returns bool. Min/Max/Floor/Ceiling return (key,value,bool).
- Single heap Push and Pop are O(log n); bulk Push rebuilds in O(n+k).
- Keys/Values/Entries return detached shallow slices. Referenced objects remain
  shared. No synchronization is added; never copy a populated container.

## Traversal and JSON

Search trees expose node-based bidirectional iterators; heap iterators traverse
forward by array index. Invalid Key/Value/Index access panics. Every successful
Put, nonempty Push/Pop, successful Remove, nonempty Clear and JSON replacement
invalidates existing iterators and Seq results, including value-only updates.
Begin/End reset position, not validity. Failed lookups/removals and empty Clear
do not invalidate traversal. Callback mutations are checked on early termination.

Each, Any, All, Find, Map and Select use the local container conventions. Find
returns bool; a heap miss also returns index -1. Map/Select preserve the
concrete structure, comparator and B-tree order. Nil callbacks panic.

Search-tree JSON is an entry array: [{"key":1,"value":"one"}]. Heap JSON is a
value array. Empty containers encode as []; null input clears. Decoding uses a
temporary structure, including heap rebuilding, before replacing the original.
Comparators are retained; malformed input leaves existing contents unchanged.
Comparator panics propagate; no rollback is promised for regular mutations
whose comparison callback panics. Identity-bearing objects are newly decoded
objects, not restored original pointer identities.

## Verification

Tests cover randomized model comparison, key bounds, natural/custom constructors,
non-comparable keys, iterator invalidation, JSON failure atomicity, and min/max
heap extraction. Internal tests validate AVL height balance, red-black root/red
adjacency/black heights, and B-tree occupancy/parent links/equal leaf depth across
orders 3, 4, 5, 8 and 32. Fuzzing exercises search-tree operation sequences.
