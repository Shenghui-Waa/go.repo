// Package maps defines associative containers with explicit key semantics.
package maps

import (
	"fmt"

	"github.com/Shenghui-Waa/go.repo/containers"
)

// Entry is a shallow snapshot and the JSON wire representation of a mapping.
type Entry[K, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value"`
}

// Map does not require Go-comparable keys. Implementations choose hash/equality
// or comparator equivalence. Ordinary values may be any type.
type Map[K, V any] interface {
	containers.Container[V]
	Put(K, V)
	Get(K) (V, bool)
	Remove(K) bool
	Keys() []K
}

// Membership is optional key/value membership testing.
type Membership[K, V any] interface {
	ContainsKey(K) bool
	ContainsValue(V) bool
}

type EntriesProvider[K, V any] interface {
	Entries() []Entry[K, V]
}

// ExtendedMap preserves the former all-capabilities Map contract.
type ExtendedMap[K, V any] interface {
	Map[K, V]
	Membership[K, V]
	EntriesProvider[K, V]
	containers.Clearer
	containers.EnumerableWithKey[K, V]
	containers.KeyedSequence[K, V]
	containers.JSONCodec
	fmt.Stringer
}

type ExtendedBidiMap[K, V any] interface {
	BidiMap[K, V]
	ExtendedMap[K, V]
}

// BidiMap maintains a one-to-one relation; inserting a duplicate value evicts
// its previous owner, matching GoDS's last-write-wins policy.
type BidiMap[K, V any] interface {
	Map[K, V]
	GetKey(V) (K, bool)
	RemoveValue(V) bool
}
