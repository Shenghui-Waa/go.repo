# Optional capabilities / 可选能力

Container[T] now requires only Values() []T, Size() int and Empty() bool.
It does not require mutation, formatting, serialization, equality or hashing.
Container[T] 现在仅要求 Values、Size、Empty，不强制清空、格式化、序列化或对象契约。

| Capability / 能力 | Interface / 接口 |
| --- | --- |
| Clear / 清空 | containers.Clearer |
| Diagnostics / 字符串表示 | fmt.Stringer |
| JSON output/input / JSON 输出与输入 | containers.JSONSerializer / JSONDeserializer |
| Both JSON directions / 双向 JSON | containers.JSONCodec |
| Indexed/keyed sequence / 索引与键值序列 | containers.IndexedSequence / KeyedSequence |
| Callbacks / 回调枚举 | containers.EnumerableWithIndex / EnumerableWithKey |
| Equality and hash / 相等与哈希 | base.Equaler / base.Hasher |
| List membership / 列表查找 | lists.Searchable |
| List sorting / 列表排序 | lists.Sortable |
| Linked-list endpoints / 链表首尾操作 | lists.Endpoints |
| Map membership / 映射查找判断 | maps.Membership |
| Entry snapshots / 键值条目副本 | maps.EntriesProvider |

List retains Get, Add, Insert, Set and Remove, plus Container. Map retains Put,
Get, Remove and Keys, plus Container. Set retains Add/Remove/Contains. Queue and
Stack retain their enqueue/dequeue or push/pop operations and Peek. Tree remains
an observation-only Container. Base interfaces do not restrict element equality
policies; concrete implementations continue to enforce their documented policies.

List 保留 Get、Add、Insert、Set、Remove；Map 保留 Put、Get、Remove、Keys；
Set 保留 Add、Remove、Contains；Queue/Stack 保留入队出队或入栈出栈及 Peek。
这些接口均组合只读 Container；Tree 仍仅组合 Container。具体容器原有方法及
哈希/相等性约束仅用于需要该能力的操作；列表基础存取现已允许任意类型。

List storage now accepts arbitrary types. Default lookup and object operations
retain runtime element checks; ObjectList expresses the optional default object
contract. EqualFunc/HashFunc and IndexOfFunc/ContainsFunc accept explicit policies.
列表的 ObjectList 不是元素可哈希性的静态证明；使用显式函数策略可避免隐式判等选择。

## Migration / 迁移

All concrete methods remain available. Code holding a concrete *List or *Map
does not need changes. Code using the old broad interfaces to call optional
methods must explicitly compose the capabilities it needs. This is an intentional
source-level interface change, not a promise of unchanged interface callers.

具体容器方法全部保留。使用具体指针类型的调用方式不变；通过原宽接口调用
可选方法的代码需要调整声明。这是有意的源码级接口变更。

```go
type JSONList[T any] interface {
    lists.List[T]
    containers.JSONSerializer
}

type ClearableMap[K, V any] interface {
    maps.Map[K, V]
    containers.Clearer
}
```

For migration convenience, ExtendedList, ExtendedMap, ExtendedBidiMap,
ExtendedSet, ExtendedQueue and ExtendedStack preserve the previous full contracts.
Prefer minimal/local compositions in new APIs. Existing Tree callers that need
Clear or String should add Clearer or fmt.Stringer explicitly.

为迁移提供 ExtendedList、ExtendedMap、ExtendedBidiMap、ExtendedSet、
ExtendedQueue、ExtendedStack，保留此前完整契约；新 API 优先选择最小组合。
原 Tree 调用方若需要 Clear/String，请显式组合 Clearer/fmt.Stringer。

Set algebra accepts basic Set operands without Seq, by reading their Values
snapshot. List object equality requires the other list to also provide Equaler
and Hasher, but does not require JSON or enumeration. A minimal third-party list
without an object-equality contract is not implicitly treated as value-equal.

集合运算通过 Values 支持没有 Seq 的基础 Set。列表对象相等比较仅额外要求
对方具备 Equaler/Hasher，不依赖 JSON 或枚举能力；没有对象相等契约的第三方
基础列表不会自动被视为内容相等。
