// Package iterable defines iterator contracts shared by the containers.
package iterable

import (
	"errors"
	"iter"
)

var (
	ErrNoSuchElement          = errors.New("no such element")
	ErrIllegalState           = errors.New("illegal iterator state")
	ErrUnsupportedOperation   = errors.New("unsupported operation")
	ErrConcurrentModification = errors.New("concurrent modification")
)

type Iterable[T any] interface {
	All() iter.Seq[T]
}

type Iterator[T any] interface {
	HasNext() bool
	Next() (T, error)
	Remove() error
	Close()
}

// SeqIterator adapts a Go sequence to a Java-shaped pull iterator.
type SeqIterator[T any] struct {
	next      func() (T, bool)
	stop      func()
	value     T
	buffered  bool
	exhausted bool
}

func FromSeq[T any](sequence iter.Seq[T]) *SeqIterator[T] {
	next, stop := iter.Pull(sequence)
	return &SeqIterator[T]{next: next, stop: stop}
}

func (it *SeqIterator[T]) HasNext() bool {
	if it.buffered {
		return true
	}
	if it.exhausted {
		return false
	}
	value, ok := it.next()
	if !ok {
		it.Close()
		return false
	}
	it.value = value
	it.buffered = true
	return true
}

func (it *SeqIterator[T]) Next() (T, error) {
	if !it.HasNext() {
		var zero T
		return zero, ErrNoSuchElement
	}
	value := it.value
	var zero T
	it.value = zero
	it.buffered = false
	return value, nil
}

func (it *SeqIterator[T]) Remove() error {
	return ErrUnsupportedOperation
}

func (it *SeqIterator[T]) Close() {
	if it.exhausted {
		return
	}
	it.exhausted = true
	it.stop()
	var zero T
	it.value = zero
	it.buffered = false
}
