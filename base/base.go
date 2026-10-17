package base

import "fmt"

// Equaler and Hasher must agree: equal values must have equal hash codes.
type Equaler interface {
	Equal(other any) bool
}

type Hasher interface {
	HashCode() uint64
}

// CompareInterface is retained for compatibility; it does not define ordering.
type CompareInterface interface {
	Equaler
	Hasher
}

type StringerInterface interface {
	ToString() string
}

// Base provides pointer identity, not Java-style inheritance. Embedded types
// needing value equality must implement both Equal and HashCode themselves.
// The nonzero size ensures distinct live allocations have distinct addresses.
type Base struct{ marker byte }

// BaseIdentity exposes the identity carried by an embedded Base. Embedding a
// Base value gives each containing object its own identity; embedding *Base
// deliberately shares identity when the same pointer is reused.
func (b *Base) BaseIdentity() *Base { return b }

func (b *Base) Equal(other any) bool {
	if IsNil(other) {
		return b == nil
	}
	value, ok := other.(interface{ BaseIdentity() *Base })
	return ok && b == value.BaseIdentity()
}

func (b *Base) HashCode() uint64 {
	return hashComparable(b)
}

func (b *Base) String() string {
	if b == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T@%x", b, b.HashCode())
}

// ToString is the Java-style alias of String.
func (b *Base) ToString() string { return b.String() }

// New returns a distinct object with pointer identity.
func New() *Base {
	return &Base{}
}
