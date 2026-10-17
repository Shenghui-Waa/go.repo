package containers

import "iter"

// Clearer optionally supports removing all elements.
type Clearer interface{ Clear() }

// IndexedSequence optionally exposes traversal indices and values.
type IndexedSequence[T any] interface{ Seq() iter.Seq2[int, T] }

// KeyedSequence optionally exposes keys and values.
type KeyedSequence[K, V any] interface{ Seq() iter.Seq2[K, V] }

// JSONCodec combines the independent serialization and deserialization contracts.
// Callers needing only one direction should use JSONSerializer/JSONDeserializer.
type JSONCodec interface {
	JSONSerializer
	JSONDeserializer
}
