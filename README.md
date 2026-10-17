# go.repo

[Chinese README](./README_zh.md)

A Go 1.25 generic data-structure library inspired by GoDS and selected Java 25
collection contracts, using Go composition, typed functions and iterators.
It is not a drop-in replacement for either library. This document describes the
current checkout, not the feature set of a previously published release.

**No external Go module dependencies.** Search-tree kernels are locally adapted
from GoDS source with attribution preserved; maps and sets use the local trees.

## Usage

Requires Go 1.25 or later. The canonical module and import paths declared in
`go.mod` are:

- `github.com/Shenghui-Waa/go.repo` *(recommended)*
- `gitee.com/Shenghui-Waa/go.repo`

The current project version is `v1.0.3`.

## Packages

| Package | Implementations / responsibilities |
| --- | --- |
| `base` | Optional object identity, Equaler/Hasher, nil checks, deep equality, hashing, ToString |
| `compare` | Comparable, Comparator, ordered/time comparison, composition and nil ordering |
| [containers](./containers/README.md) | Shared interfaces, sorted snapshots, sequences and JSON contracts |
| [lists](./lists/README.md) | arraylist, singlylinkedlist, doublylinkedlist |
| [maps](./maps/README.md) | hashmap, linkedhashmap, treemap, hashbidimap, treebidimap |
| [trees](./trees/README.md) | avltree, redblacktree, btree, binaryheap |
| [sets](./sets/README.md) | hashset, linkedhashset, treeset |
| [queues](./queues/README.md) | arrayqueue, linkedlistqueue, circularbuffer, priorityqueue |
| [stacks](./stacks/README.md) | arraystack, linkedliststack |

There is no catch-all utils package. Comparison belongs in compare, object
conversion in base, and container-specific behavior in its owning package.

## Quick example

```go
package main

import (
    "fmt"

    "github.com/Shenghui-Waa/go.repo/compare"
    "github.com/Shenghui-Waa/go.repo/lists/arraylist"
    "github.com/Shenghui-Waa/go.repo/maps/treemap"
    "github.com/Shenghui-Waa/go.repo/queues/priorityqueue"
    "github.com/Shenghui-Waa/go.repo/sets/hashset"
    "github.com/Shenghui-Waa/go.repo/stacks/arraystack"
)

func main() {
    list := arraylist.New[int]()
    list.Add(3, 1, 2)
    list.Sort(compare.Ordered[int]())
    fmt.Println(list.Values()) // [1 2 3]

    mapping := treemap.NewWithComparator[int, string](compare.Ordered[int]())
    mapping.Put(2, "two")
    fmt.Println(mapping.Get(2)) // two true

    set := hashset.New[int]()
    fmt.Println(set.Add(1, 1, 2)) // 2 new members

    queue := priorityqueue.NewWithComparator(compare.Ordered[int]())
    queue.Enqueue(3)
    queue.Enqueue(1)
    fmt.Println(queue.Dequeue()) // 1 true

    stack := arraystack.New[string]()
    stack.Push("first")
    stack.Push("last")
    fmt.Println(stack.Pop()) // last true
}
```

## Construction and ordering

Public collection interfaces expose only core operations. Formatting, clearing,
JSON, enumeration and object equality/hash are optional capabilities; concrete
implementations retain their methods. See [capability composition and migration](./containers/CAPABILITIES.md).
ExtendedList/ExtendedMap and corresponding Extended* interfaces preserve the old
full contracts for callers that need them.

- All public New constructors are parameterless, with explicit generic types
  as needed. Insert initial elements through mutation methods.
- Lists, hash maps/sets, ordinary queues and stacks do not require CompareTo.
- Tree maps/sets, search trees, heaps and priority queues require ordering.
  Their New requires `CompareTo(T) int`; NewWithComparator accepts the local
  Comparator. Tree bidimaps require orders for both keys and values.
- Built-in types use `compare.Ordered[T]()`, natural orders use
  `compare.FromComparable[T]()`, and time.Time uses `compare.Time()`.
- Reversed, ThenComparing, Comparing, ComparingOrdered, NilFirst and NilLast
  compose orders. Lists accept a saved comparator or an explicit Sort argument.
- B-tree default order is 32. NewWithOrder and NewWithOrderAndComparator support
  orders at least 3. Circular buffers default to capacity 16; NewWithCapacity
  requires a positive capacity. Full buffers overwrite the oldest value.

## Equality and identity

Tree indexes identify comparator-zero keys; hash indexes use HashCode plus Equal.
Tree updates retain the original key representative. Keep ordering consistent
with equality when mixing containers, and do not mutate indexed fields in place.

List storage accepts arbitrary types. Its default value lookup and object
Equal/HashCode methods, and hash-indexed operands, accept built-in/named scalars or objects
implementing both base.Equaler and base.Hasher. Objects may embed base.Base and
use pointers for default identity, or implement both methods directly. Overriding
Equal requires a consistent HashCode override; Go cannot prove this agreement.
Base composition is not Java inheritance or virtual dispatch.

For arbitrary list elements, use lists.EqualFunc/HashFunc and
IndexOfFunc/ContainsFunc with explicit policies. ObjectList exposes the optional
default object contract; it does not statically guarantee hashable contents.

Ordinary map values, FIFO/LIFO elements and comparator-only operands may be any
type. Ordinary Map.ContainsValue needs equality-capable operands; use Any with
a custom predicate otherwise. Hash bidimaps need hash/equality on both sides.
Native NaN hash keys/members are rejected. Nil support follows element contracts;
ordered nil operands require a suitable comparator.

base.HashCode hashes one value; base.Hash combines values in order into uint64,
without guaranteeing Java-compatible numeric hashes. DeepEquals is separate
reflection-based comparison, not the ordinary hashing contract. base.ToString
normalizes nil-like values to `<nil>`, prefers Go formatting protocols, then
supports ToString-only objects. A promoted Base.String takes precedence over an
outer ToString-only override. Existing container formatting is unchanged.

## Operations, traversal and JSON

- Lists retain duplicates; invalid indexed mutations return false. Maps overwrite
  equivalent keys; bidimaps evict an incoming value's previous owner. Set Add and
  Remove return counts. Retrieval uses success flags to distinguish zero values.
- Set algebra uses receiver equivalence. Mixed policies may yield asymmetric
  results; use matching policies for ordinary mathematical set identities.
- Snapshots are shallow. Seq/iterators check mutation versions; resetting cursor
  position cannot restore invalidated traversal. Allocation and invalidation
  details are documented per package. These checks are not synchronization.
- Each/Any/All/Find/Map/Select reject nil callbacks. Any/All/Find short-circuit;
  empty Any is false and empty All is true. Transformations retain concrete
  types; containers.Map supports changing the output element type.
- Lists/sets/queues/stacks/heaps encode JSON value arrays; maps/search trees use
  `[{"key":...,"value":...}]`. Empty containers encode as []; null clears.
  Decode uses temporary storage before committing, retaining order/capacity settings.
- Stack JSON/traversal is top-first. Heap/priority queue traversal is heap layout,
  not sorted order. Circular decoding beyond capacity retains newest values.

Containers are not concurrent-safe; never copy populated instances. User
comparison/hash/equality callbacks must be pure, deterministic and non-panicking.
Regular mutations do not promise rollback after callback failure. General cyclic
object serialization/stringification is not supported.

## Validation

Run from the module root:

```sh
go test -mod=readonly ./...
go vet ./...
go test ./maps -run '^$' -fuzz FuzzMaps -fuzztime=5s
go list -m all
```

Tests cover randomized models, JSON/iterator contracts, collisions, tree
invariants and focused fuzz targets. With no external dependencies, go list -m all
should list only this module. go mod tidy is safe: GoDS and its go.sum entries
are no longer intentionally retained.

## License and attribution

Original project code uses [MIT](./LICENSE). Adapted GoDS tree kernels retain
their notices and applicable BSD/ISC terms in [trees/LICENSE.GoDS](./trees/LICENSE.GoDS).
Removing the dependency does not remove source-license obligations.
