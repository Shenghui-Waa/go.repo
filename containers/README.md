# Container foundations

This package is a fresh implementation of the GoDS container capabilities,
using Go 1.25 generics and the local `compare` package. It does not change the
reference `gods/` checkout. The project has no external Go module dependencies.
It defines contracts and shared helpers, not concrete lists, maps or trees.

## Design

- `Container[T]` only requires Values, Size and Empty. Clear is the optional
  Clearer capability; diagnostics use the standard fmt.Stringer interface.
- Sorting uses `compare.Comparator[T]` and makes a shallow copy before a stable
  sort, even when the source exposes internal storage.
- `GetSortedComparableValues` integrates `compare.Comparable[T]` directly.
- Enumerable interfaces retain Each, Any, All and Find. Find now includes a
  success flag. Indexed methods return -1 on a miss; the generic sequence Find
  helper returns zero keys on a miss, so indexed implementations must translate
  that key to -1 when the success flag is false.
- Indexed/keyed and bidirectional iterator contracts remain separate capabilities.
  Invalid cursor access panics; structural mutation invalidates a live cursor
  unless the implementation explicitly offers other guarantees.
- `Indexed` and `Keyed` consume existing cursors without resetting them.
  `Snapshot` instead produces a repeatable shallow snapshot captured at creation.
- Sequence helpers Each, Find, Any, All, Map and Filter support early termination.
  Map can change the result type. Map and Filter are lazy; no container type or
  allocation strategy is imposed on future concrete implementations.
- JSON interfaces embed the standard library interfaces and retain ToJSON and
  FromJSON. Concrete decoders must commit only after successful decoding.

Java 25's Collection and Iterator APIs are design references for separating
container and traversal responsibilities. This is not a Java compatibility
layer: there is no universal inheritance root, mandatory hash/equality contract,
optional-operation exception, or implied fail-fast mutation detector. Future
concrete containers must choose their own equality and hash policies; `base.Base`
pointer identity is not automatically appropriate for collection value equality.

## Example

```go
// source implements containers.ValuesProvider[Task].
ordered := containers.GetSortedValuesFunc(source, compare.FromComparable[Task]())

seq := containers.Snapshot(source)
selected := containers.Filter(seq, func(index int, task Task) bool {
    return task.Active
})
names := slices.Collect(containers.Map(selected, func(index int, task Task) string {
    return task.Name
}))
```

Examples assume an application-defined Task with CompareTo(Task), Active and
Name. Snapshot and sorting are shallow: pointed-to values remain shared.
No helper adds concurrency synchronization. Callback mutation of live sources
is not supported unless the source explicitly documents it.

References:
- Java 25 Collection: https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/Collection.html
- Java 25 Iterator: https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/Iterator.html
- Go iteration: https://pkg.go.dev/iter
