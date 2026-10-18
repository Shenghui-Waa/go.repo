package lists

// Endpoints is an optional first/last capability, independent of List and
// equality/hash contracts. Empty reads/removals return (zero, false).
// Complexity depends on storage: singly linked tail removal is O(n).
type Endpoints[T any] interface {
	AddFirst(T)
	AddLast(T)
	GetFirst() (T, bool)
	GetLast() (T, bool)
	RemoveFirst() (T, bool)
	RemoveLast() (T, bool)
}
