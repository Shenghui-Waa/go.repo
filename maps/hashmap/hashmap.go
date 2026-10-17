// Package hashmap provides custom-equality hash keys with collision buckets.
package hashmap

import "github.com/Shenghui-Waa/go.repo/maps/internal/mapcore"

type Map[K, V any] struct {
	mapcore.Core[K, V, mapcore.Hash]
}
type Iterator[K, V any] = mapcore.Cursor[K, V]

func New[K, V any]() *Map[K, V] { return &Map[K, V]{} }
func (m *Map[K, V]) Map(f func(K, V) (K, V)) *Map[K, V] {
	return &Map[K, V]{m.Core.Map(f)}
}
func (m *Map[K, V]) Select(f func(K, V) bool) *Map[K, V] {
	return &Map[K, V]{m.Core.Select(f)}
}
