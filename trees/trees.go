// Package trees defines local ordered trees and priority heaps.
package trees

import "github.com/Shenghui-Waa/go.repo/containers"

type Tree[V any] interface{ containers.Container[V] }

// Entry is a detached shallow key/value snapshot, also used for JSON.
type Entry[K, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value"`
}
