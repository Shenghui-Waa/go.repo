# Stacks

arraystack and linkedliststack provide local LIFO storage with New[T]() and usable
zero values. The array implementation uses a slice; the linked implementation
uses top-linked nodes. They do not depend on list equality rules or external
GoDS packages. Original gods/ sources remain; the external module is not required.

```go
stack := arraystack.New[[]int]()
stack.Push([]int{1})
stack.Push([]int{2})
value, ok := stack.Pop() // [2], true
```

Any element type, duplicates and nil are supported. CompareTo, Equal, HashCode
and Base embedding are not required because stack operations use none of them.
Pop/Peek return (zero,false) when empty, distinguishing stored nil/zero values.
Array Push is amortized O(1); Pop/Peek are O(1). Linked Push/Pop/Peek are O(1).
Popped array slots are cleared to release referenced objects; Clear releases
storage. Do not copy populated stacks. No concurrency safety is provided.

## Ordering

Values, Seq, Iterator, callbacks and JSON all use TOP-TO-BOTTOM order. Index 0
means the top. Values returns an independent shallow slice. Map invokes its
callback in that order and preserves the transformed positions. Select preserves
the relative order of retained values. Both return the same concrete stack kind.

JSON is a top-first array, empty as []. Decoding [3,2,1] makes 3 the next Pop.
Decoding builds temporary storage before committing; malformed input preserves
the original. null clears the stack. This explicit wire-order contract guarantees
roundtrips preserve Pop order. Interface decoding follows encoding/json rules.

## Traversal

Seq and Iterator capture O(n) shallow snapshots and source versions; iterator
steps are O(1). Both stacks support bidirectional snapshot traversal. Referenced
element objects remain shared. Push, successful Pop, nonempty Clear and JSON
replacement invalidate traversal; Peek, empty Pop and empty Clear do not.
Begin/End reset position, not validity. Invalid Value/Index access panics.
Callback mutation is detected even on early termination.

Each/Any/All/Find/Map/Select reject nil callbacks. Any/All/Find short-circuit.
Empty Any is false and empty All is true. Find misses return (-1,zero,false).
Stacks intentionally do not supply automatic object equality/hash semantics.

Tests cover model-based random LIFO operations, duplicates, nil/non-comparable
values, snapshots, iterator boundaries/invalidation, transformations, JSON order
and failed decoding without mutation.
