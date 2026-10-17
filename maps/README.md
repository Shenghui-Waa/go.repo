# Maps

Five typed map implementations share the local containers contracts. The
reference gods/ checkout is unchanged. Tree indexes use the project's generic
trees/redblacktree directly, without boxed keys or runtime payload assertions.
The project-level GoDS module requirement has been removed.

| Package | Constructor | Key identity | Iteration |
| --- | --- | --- | --- |
| hashmap | New[K,V]() | HashCode + Equal | Unspecified |
| linkedhashmap | New[K,V]() | HashCode + Equal | First insertion |
| hashbidimap | New[K,V]() | HashCode + Equal on both sides | Unspecified |
| treemap | New[K,V]() / NewWithComparator[K,V](keys) | Comparator zero | Key order |
| treebidimap | New[K,V]() / NewWithComparator(keys,values) | Comparator zero on both sides | Key order |

All constructors create empty maps. Hash maps also support their zero value.
Tree maps must be constructed with an order: their zero value is not ready for
lookup or mutation. Natural-order tree New requires K to implement
CompareTo(K) int; treebidimap also requires V to implement CompareTo(V) int.
Built-in types use explicit comparators:

```go
plain := hashmap.New[string, []int]()
plain.Put("numbers", []int{1, 2})

ordered := treemap.NewWithComparator[int, string](compare.Ordered[int]())
ordered.Put(2, "two")

both := treebidimap.NewWithComparator(compare.Ordered[int](), compare.Ordered[string]())
both.Put(1, "one")
```

## Key and value contracts

Hash-indexed operands accept nil, built-in/named scalar types, or objects
implementing both base.Equaler and base.Hasher. A struct may embed base.Base
and be passed as a pointer, or implement both methods directly. Slices/maps/plain
structs without these contracts are rejected. Hash collisions are resolved by
Equal, not by hash equality. Native float/complex NaN hash keys are rejected
because native equality is not reflexive. Typed nil objects satisfying both
contracts are supported. Equal keys must have equal hashes.

Tree-indexed operands only need the comparator: comparator zero means the same
key, even when object Equal would be false. Replacing a key retains its original
stored key object. Comparators should be consistent with object equality, and
must be pure, deterministic and non-panicking. Natural-order nil handling is the
responsibility of CompareTo; use an explicit nil-aware comparator when needed.

Ordinary map values may be any type because Put/Get do not need value equality.
ContainsValue on an ordinary map requires compared values to satisfy the scalar
or Equal/HashCode contract and may panic for unsupported values. For arbitrary
values use Any with a caller-supplied predicate instead. Bidirectional maps index
their values, so hashbidimap requires hash/equality on both sides and treebidimap
requires comparators on both sides.

Overriding Equal requires a matching HashCode override; Go cannot prove that
promoted methods and overrides agree. Do not mutate indexed key fields, or
indexed value fields in bidirectional maps, while they are stored. Do not copy
maps after use. All maps are non-concurrent. Callback panics/reentrant mutation
during hashing/comparison are unsupported and do not promise rollback.

These maps intentionally do not define map-level Equal/HashCode: arbitrary
ordinary values and comparator equivalence do not provide a universal compatible
object hash contract. Element-level requirements apply only where needed.

## Behavior

- Put replaces a key's value. Remove returns a success flag. Get/GetKey/Find
  return a bool so zero values are distinguishable from absence.
- Bidirectional Put removes both the old value for the key and any other owner
  of the incoming value. The incoming pair wins. RemoveValue updates both sides.
- Linked maps retain position on replacement; removal followed by insertion
  appends. Hash maps currently traverse their internal chain but do not promise
  insertion order; use linkedhashmap when order is part of the contract.
- Keys, Values and Entries are independent shallow slices in traversal order.
  Treebidimap Values follows KEY order, unlike GoDS's separately value-sorted
  Values. Entries/Seq always provide aligned pairs.
- Tree maps expose Min, Max, Floor and Ceiling. Lookup/update/removal are O(log n).
  Hash lookup is expected O(1), degrading to O(n) for a collision bucket.
- Each, Any, All, Find, Map and Select reject nil callbacks. Any/All/Find stop
  early. Map/Select retain concrete map kind and comparators. Transformed key or
  value conflicts use the same last-write-wins rules as Put.
- Seq traverses directly. Iterator captures O(n) shallow entries and gives O(1)
  forward/backward steps. Both are fail-fast: successful Put, Remove, nonempty
  Clear and JSON replacement invalidate them. Begin/End do not restore validity.
  Invalid cursor Key/Value access panics. No thread-safety is implied.

## JSON

All maps encode an array of entries, not a JSON object:

```json
[{"key":1,"value":"one"},{"key":2,"value":"two"}]
```

This preserves typed non-string keys without string conversion. It is an
intentional wire-format difference from GoDS. Empty maps encode as []; null input
clears the map. Input entries are applied in order with Put's duplicate policy.
Decode and contract validation use temporary indexes before committing, preserving
the original on failure. Comparators survive replacement. Object identity keys
are newly allocated on decoding, not identities of the pre-serialization objects.
Interface decoding follows encoding/json and does not infer custom Go types.
Malformed JSON returns errors; unsupported hash keys return contract errors;
panics from user comparators are not swallowed. Cyclic values are unsupported
by JSON, and user String methods must not recurse through the containing map.

Reference: Java 25 TreeMap comparator-equivalence design:
https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/TreeMap.html
