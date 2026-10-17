# Lists

The three packages implement insertion-ordered, duplicate-preserving lists:

| Package | Storage | Get/Set | Append | Forward traversal |
| --- | --- | --- | --- | --- |
| arraylist | Go slice | O(1) | Amortized O(1) | O(n) |
| singlylinkedlist | Singly linked nodes with tail | O(n) | O(1) | O(n) |
| doublylinkedlist | Doubly linked nodes with head/tail | O(min(i,n-i)) | O(1) | O(n) |

Indexed insertion/removal additionally requires shifting array elements or
locating linked nodes. All implementations support their zero value. Do not copy
a list after first use. Lists and iterators are not concurrency-safe.

## Construction and ordering

```go
values := arraylist.New[int]()
values.Add(3, 1, 2)
values.Sort(compare.Ordered[int]())

descending := doublylinkedlist.NewWithComparator(compare.Ordered[int]().Reversed())
descending.Add(1, 3, 2)
descending.Sort()

people := singlylinkedlist.New[Person]() // No CompareTo requirement.
// When Person implements CompareTo(Person) int:
people.Sort(compare.FromComparable[Person]())
```

New[T]() is strictly parameterless. NewWithComparator takes a non-nil local
compare.Comparator[T]. Sort() uses the saved comparator; Sort(order) overrides
it for that operation without changing the saved default. No automatic natural
order is inferred: use FromComparable explicitly. Missing/nil comparators panic,
even on empty lists. Sorting is stable and works on a snapshot before committing;
a comparator panic does not partially reorder the original list. Comparators
must not mutate the list or the fields involved in ordering.

## Equality and Base

Storage accepts any T: Add, Insert, Set, indexed removal, traversal, transformation
and JSON decoding do not require Equal or HashCode. Sorting requires only its
explicit comparator. Plain structs, slices, maps and functions can be stored
(JSON support itself still follows encoding/json).

Default Contains/IndexOf/LastIndexOf/RemoveValue and object Equal/HashCode retain
their scalar-or-Equaler+Hasher policy. Unsupported searched or compared operands
panic only when these operations are explicitly invoked, not on insertion.
Short-circuiting does not certify the remaining elements. ObjectList names this
optional object contract; ExtendedList retains it for compatibility. Since the
existing concrete types still expose these methods, Go interface satisfaction
alone does NOT guarantee all stored elements are hashable.

For arbitrary values use explicit policies, without reflection-based defaults:

```go
items := arraylist.New[[]int]()
items.Add([]int{1, 2})
found := lists.ContainsFunc[[]int](items, []int{1, 2}, slices.Equal[[]int, int])
```

lists.EqualFunc compares ordered snapshots, lists.HashFunc combines supplied
element hashes, and lists.IndexOfFunc/ContainsFunc use explicit equality. Nil
callbacks panic even for empty lists. Callbacks must not mutate their sources.
Equal elements must have equal hashes when these two policies are used together.
No policy is silently inferred from a sorting comparator.

中文：基础存取、遍历及 JSON 解码支持任意 T，不要求 Base 或 Equal/HashCode。
默认值查找、Equal、HashCode 仅在调用时执行既有对象契约检查；它们仍可能 panic。
ObjectList 显式表示这项能力，但不是元素可哈希性的编译期证明。切片、映射等
可使用 EqualFunc、HashFunc、IndexOfFunc、ContainsFunc 提供明确策略；查找无需哈希。

```go
type IdentityItem struct {
    base.Base
    Name string
}

items := arraylist.New[*IdentityItem]()
item := &IdentityItem{Name: "example"}
items.Add(item)
// items.Contains(item) is true; another allocation has a different identity.

type Person struct { ID int }
func (p Person) Equal(other any) bool {
    q, ok := other.(Person)
    return ok && p.ID == q.ID
}
func (p Person) HashCode() uint64 { return base.HashCode(p.ID) }
```

Base methods have pointer receivers: use *IdentityItem when embedding a Base
value. BaseIdentity makes promoted equality recognize the containing object's
identity. Embedding *Base requires initialization; sharing that pointer shares
identity. Avoid nil embedded *Base and copying identity-bearing objects after use.

Overriding Equal requires overriding HashCode consistently. Go's method sets
cannot tell an explicitly overridden method from a promoted method reliably;
the library checks both methods exist, but cannot prove their semantic agreement.
Equality must be symmetric/transitive. Comparator equality is never used by
Contains, IndexOf or RemoveValue. Float NaN follows native unequal-to-itself rules.

Lists implement ordered element-wise Equal and HashCode across the three list
types with the same T when elements satisfy the default policy. Hashing combines
element hashes in order. Nil elements can always be stored; default object methods
still check their declared/dynamic type contracts. Avoid self-referential lists in Equal, HashCode, String or JSON
operations: recursive cycles are not supported. Mutable list hashes must not be
used as fixed hash-table keys while their contents are changing.

## Operations and compatibility

- Add/Append append, Prepend inserts at the beginning, and Insert accepts [0,Size].
- Get returns (value, bool). Remove, RemoveValue, Set, Swap and Insert return bool.
  Invalid indices return false without mutation. Set(Size(), value) does not append.
- Contains() is true; RemoveValue removes the first match. IndexOf/LastIndexOf
  return -1 for a miss. Values returns an independent shallow slice, empty as [].
- Find returns (index, value, bool), with index -1 for a miss. Any/All short-circuit.
- Each, Map and Select reject nil callbacks, including on empty lists. Map/Select
  return the same concrete list kind and preserve the saved comparator. For a
  different output type, use containers.Map on Seq().
- Iterator is bidirectional for array/double lists and forward-only for single
  lists. Invalid Index/Value access panics. Begin/End reset position, not validity.
- Successful changes, including Set, Swap of distinct indices, Sort of two or
  more values, and JSON replacement, invalidate existing cursors and Seq results.
  No-op empty insertions, Clear on empty and Swap(i,i) do not invalidate them.
  Callback mutation is detected even when the callback requests early termination.
- JSON methods encode values as an array; empty lists encode as []. Input null
  clears the list. Decoding finishes before replacing data; it does not validate hashing.
  The comparator remains unchanged. JSON of interface elements follows the
  standard decoder and does not reconstruct arbitrary custom object types.

The implementation borrows Java List's separation of sorting and equality and
ordered list equality/hash semantics, without Java exceptions or inheritance.
It retains GoDS operation names where useful but is not a drop-in GoDS API.
The original gods/ checkout is unchanged; no external GoDS module is required.

Reference: https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/List.html
