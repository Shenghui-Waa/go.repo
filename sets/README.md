# Sets

Three generic sets use this project's maps, including the local red-black tree
behind treemap. No external GoDS packages are imported by the set dependency
chain. The project-level GoDS requirement is removed; original gods/ sources remain.

| Package | Constructor | Membership | Traversal |
| --- | --- | --- | --- |
| hashset | New[T]() | HashCode + Equal | Unspecified |
| linkedhashset | New[T]() | HashCode + Equal | First insertion |
| treeset | New[T]() / NewWithComparator(order) | Comparator zero | Comparator order |

```go
hash := hashset.New[int]()
hash.Add(3, 1, 3) // Returns 2.

linked := linkedhashset.New[string]()
linked.Add("first", "second")

ordered := treeset.NewWithComparator(compare.Ordered[int]())
ordered.Add(3, 1, 2)
```

Constructors take no initial elements. Tree New[T]() requires CompareTo(T) int;
use NewWithComparator with built-in values or a custom ordering. Nil comparators
are rejected. Hash and linked sets support their zero values; tree sets require
a constructor. Sets are not concurrency-safe and must not be copied after use.

## Membership

Hash sets follow local map contracts: built-in/named scalar values use native
equality; objects must implement both Equal and HashCode, directly or by embedding
Base and using a pointer. Typed nil objects satisfying both contracts are allowed.
Hash collisions are resolved by Equal. Native float/complex NaN members are
rejected. Plain structs, raw pointers, slices and maps without the contract are
rejected. Overriding Equal requires a consistent HashCode override; Go cannot
prove agreement between overrides and promoted methods.

Tree sets only require a comparator, supporting non-comparable Go values. A zero
comparison means duplicate membership even if object Equal would differ. Keep
ordering consistent with equality when interoperability matters. Do not mutate
stored fields used by hashing, equality or ordering. User callbacks must be pure,
deterministic and non-panicking; rollback is not promised for callback failures.

Add/Remove return the number of members actually added/removed. Duplicate Add
retains the original member and insertion position. Missing Remove is a no-op.
Contains() is true. Values is an independent shallow slice. Collection-level
Equal/HashCode are intentionally not supplied, because arbitrary comparator
equivalence cannot provide a universal object-hashing contract.

## Set algebra

Union, Intersection, Difference and SymmetricDifference return the receiver's
concrete set kind and retain its comparator. Both sources remain unchanged.
IsSubsetOf, IsSupersetOf and IsDisjoint use the same membership policy.

All operations interpret the other operand using RECEIVER equivalence. This
makes behavior explicit for mixed hash/tree or differently ordered tree sets:
case-insensitive A intersect case-sensitive a contains A, while the opposite
operation is empty. Algebra is not commutative across different equivalence
policies. Use matching policies for ordinary mathematical set identities.
Members unsupported by the receiver's policy cause a panic; nil operands panic.

For linked results, receiver members retain receiver order, followed by new
members from the other operand in its traversal order. Receiver representatives
win duplicate equivalence classes. Map deduplicates transformed values; Select
preserves selected representatives. Both preserve type, ordering and comparator.

## Traversal and JSON

Each, Any, All, Find, Map and Select reject nil callbacks. Any/All/Find stop early;
Find returns (index,value,bool), with index -1 on a miss. Indices are traversal
positions, not stable element identities. Empty Any is false; empty All is true.

Seq traverses the backing map directly. Iterator provides forward/reverse steps
over an O(n) shallow snapshot, but remains fail-fast against the source version.
Adding a new member, removing a member, nonempty Clear and JSON replacement
invalidate traversals. Duplicate Add, missing Remove and empty Clear do not.
Begin/End reset position, not validity. Invalid Value/Index access panics.
Callbacks that mutate membership are detected even when stopping early.

Tree sets additionally provide Min/Max/Floor/Ceiling, returning (value,bool).
Bounds are relative to the comparator, including reversed order.

JSON uses a value array, empty as [], and null input clears. Duplicate values
are deduplicated while retaining the first representative. Decode uses temporary
map indexes and commits only on success; malformed or unsupported hash operands
leave the original unchanged. Comparators survive decoding. Interface values use
standard encoding/json decoding, not automatic reconstruction of custom types.
Identity-based objects acquire new identities upon decoding.
