# go.repo

[英文说明](./README.md)

基于 Go 1.25 的泛型数据结构库，参考 GoDS 与 Java 25 的部分集合契约，采用
Go 的组合、类型安全函数与迭代器设计，并非两者的直接兼容替代品。
本文描述当前工作区实现，不代表此前已发布版本拥有全部功能。

项目搜索树算法由 GoDS 源码本地迁移适配，保留来源
与许可证；maps、sets 已使用本项目的树实现。

## 使用

需要 Go 1.25 或更新版本。go.mod 声明的规范模块及导入路径为：

- `github.com/Shenghui-Waa/go.repo`*（建议）*
- `gitee.com/Shenghui-Waa/go.repo`

项目当前版本号为 `v1.0.3`

## 模块目录

| 包 | 实现与职责 |
| --- | --- |
| `base` | 可选对象身份、Equaler/Hasher、判空、深比较、哈希、ToString |
| `compare` | Comparable、Comparator、内置/时间排序、比较器组合与 nil 排序 |
| [containers](./containers/README.md) | 公共接口、排序副本、序列工具及 JSON 契约 |
| [lists](./lists/README.md) | arraylist、singlylinkedlist、doublylinkedlist |
| [maps](./maps/README.md) | hashmap、linkedhashmap、treemap、hashbidimap、treebidimap |
| [trees](./trees/README.md) | avltree、redblacktree、btree、binaryheap |
| [sets](./sets/README.md) | hashset、linkedhashset、treeset |
| [queues](./queues/README.md) | arrayqueue、linkedlistqueue、circularbuffer、priorityqueue |
| [stacks](./stacks/README.md) | arraystack、linkedliststack |

不设统一 utils 包：比较能力归属 compare，对象转换归属 base，具体容器行为
保留在对应包。链接中的子包文档提供更详细的英文 API 说明。

## 完整示例

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
    fmt.Println(set.Add(1, 1, 2)) // 新增 2 个元素

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

## 构造与比较器

公共集合接口仅要求核心操作；格式化、清空、JSON、枚举、对象相等与哈希是可选能力，
具体实现的方法不删除。详见[能力组合与迁移说明](./containers/CAPABILITIES.md)。
ExtendedList/ExtendedMap 等 Extended* 接口保留此前完整契约，供需要这些能力的调用方迁移。

- 所有公开 New 均为空参构造，按需指定泛型；通过后续增添方法放入初始元素。
- lists、哈希 maps/sets、普通 queues 和 stacks 不要求元素实现 CompareTo。
- 树映射、树集合、搜索树、堆及优先队列需要排序；其 New 要求排序类型实现
  `CompareTo(T) int`，否则使用本项目 NewWithComparator。双向树映射两侧都需排序。
- 内置类型使用 `compare.Ordered[T]()`，自然排序适配使用
  `compare.FromComparable[T]()`，时间点比较使用 `compare.Time()`。
- Reversed、ThenComparing、Comparing、ComparingOrdered、NilFirst、NilLast
  可组合排序规则。列表可使用构造时保存或调用 Sort 时显式传入的比较器。
- B 树默认 32 阶；NewWithOrder、NewWithOrderAndComparator 支持至少 3 阶。
  循环缓冲区默认容量 16，NewWithCapacity 要求正容量；满后覆盖最旧元素。

## 相等性与对象身份

树索引以比较结果为零判定同键，哈希索引使用 HashCode 和 Equal。
树更新保留原键对象。混合使用容器时应使两种规则一致，存储期间不要修改索引字段。

列表基础存取支持任意类型。列表默认值查找、对象 Equal/HashCode 以及哈希索引
操作数要求内置/命名标量，或同时实现 base.Equaler 与 base.Hasher
的自定义对象。对象可组合 base.Base 并使用指针获得默认身份语义，也可自行实现
两种方法。重写 Equal 必须同步重写一致的 HashCode；Go 无法证明这种语义一致性。
组合 Base 不等于 Java 继承或虚方法派发。

任意列表元素可通过 lists.EqualFunc/HashFunc、IndexOfFunc/ContainsFunc 指定策略。
ObjectList 表示可选默认对象契约，不代表编译期已保证所有元素可哈希。

普通 Map 的值、FIFO/LIFO 元素及仅依赖比较器的操作数可以是任意类型。
普通 Map.ContainsValue 要求相等性能力；任意值可改用 Any 加自定义谓词。
哈希双向映射两侧都需哈希/相等性契约。原生 NaN 不可作为哈希键或哈希集合成员。
nil 支持遵循元素契约；有序结构处理 nil 需要合适的比较器。

base.HashCode 处理单值，base.Hash 按顺序组合多值并返回 uint64，不承诺与 Java
哈希数值一致。DeepEquals 是独立反射深比较，不是普通哈希的相等性契约。
base.ToString 将 nil 类值转换为 `<nil>`，优先使用 Go 标准格式化协议，再兼容仅
实现 ToString 的对象。组合 Base 后提升的 String 优先于外层仅覆盖的 ToString；
现有容器输出格式保持不变。

## 操作、遍历与 JSON

- 列表保留重复值，非法索引修改返回 false。Map 覆盖等价键；双向映射移除新值
  原来的所有者。Set.Add/Remove 返回变化数量，取值使用成功标记区分零值与不存在。
- 集合运算使用接收方判等策略，策略不同可产生非对称结果；需要普通数学集合
  性质时应采用一致策略。
- 快照为浅拷贝。Seq/迭代器检查修改版本，重置位置不恢复已失效的遍历。
  分配成本及具体失效规则见子包文档；这些检查不是并发同步。
- Each/Any/All/Find/Map/Select 拒绝 nil 回调，Any/All/Find 短路执行。
  空集合 Any 为 false、All 为 true。转换保留具体容器类型；变更输出类型可用 containers.Map。
- lists/sets/queues/stacks/heaps 使用 JSON 值数组；maps/搜索树使用
  `[{"key":...,"value":...}]`。空容器编码为 []，null 输入清空。解码先构造临时
  存储再替换，保留比较器或容量设置。
- 栈的 JSON/遍历按栈顶到栈底；优先队列/堆按堆布局遍历而非完整排序。
  循环缓冲区解码超过容量时保留最新元素。

容器不保证并发安全，不应复制已使用的实例。比较/哈希/相等性回调必须纯净、
确定且不发生 panic；普通修改操作不承诺回调失败后回滚。
不提供通用的循环对象安全序列化或字符串转换。

## 验证

在模块根目录运行：

```sh
go test -mod=readonly ./...
go vet ./...
go test ./maps -run '^$' -fuzz FuzzMaps -fuzztime=5s
go list -m all
```

测试覆盖随机模型、JSON/迭代器契约、碰撞处理、树不变量及针对性模糊测试。
无外部依赖时，go list -m all 应仅列出本项目模块。可以正常执行 go mod tidy，
不再预留外部 GoDS 依赖或对应 go.sum。

## 许可证与来源

项目原创代码采用 [MIT](./LICENSE)。从 GoDS 适配的树算法保留来源声明，以及
[trees/LICENSE.GoDS](./trees/LICENSE.GoDS) 中适用的 BSD/ISC 条款。
移除模块依赖不等于取消这些源码许可证义务。
