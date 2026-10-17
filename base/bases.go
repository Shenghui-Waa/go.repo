package base

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"math"
	"reflect"
)

// IsNil includes interfaces holding typed nil pointers, maps, slices, channels,
// functions and unsafe pointers. Other zero values are not nil.
func IsNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	default:
		return false
	}
}

// NonNil is the inverse of IsNil.
func NonNil(value any) bool { return !IsNil(value) }

// Equals treats all nil-like values as equal, then uses the left operand's
// Equaler contract, or Go equality for comparable values of the same type.
// Non-comparable values without Equaler are unequal; use DeepEquals for them.
// Custom equality must be symmetric and consistent with Hasher.
func Equals(left, right any) bool {
	if IsNil(left) || IsNil(right) {
		return IsNil(left) && IsNil(right)
	}
	if value, ok := left.(Equaler); ok {
		return value.Equal(right)
	}
	l, r := reflect.ValueOf(left), reflect.ValueOf(right)
	return l.Type() == r.Type() && l.Comparable() && r.Comparable() && l.Equal(r)
}

// DeepEquals uses reflect.DeepEqual, including its cycle handling and distinction
// between nil and empty collections. Nil-like top-level values compare equal.
// It does not invoke Equaler and is not the equality contract for HashCode.
func DeepEquals(left, right any) bool {
	if IsNil(left) || IsNil(right) {
		return IsNil(left) && IsNil(right)
	}
	return reflect.DeepEqual(left, right)
}

// HashCode returns zero for nil-like values, delegates to Hasher, or hashes a
// comparable Go value. It panics for unsupported values or an Equaler without
// Hasher. Pointer hashes are process-local, not persistent identifiers.
// Custom Equaler and Hasher methods apply only to the top-level value.
func HashCode(value any) uint64 {
	if IsNil(value) {
		return 0
	}
	if value, ok := value.(Hasher); ok {
		return value.HashCode()
	}
	if _, ok := value.(Equaler); ok {
		panic("base: Equaler must also implement Hasher")
	}
	return hashComparable(value)
}

// Hash combines zero or more values in order, like Objects.hash. The empty
// combination is 1; Hash(value) is distinct from single-value HashCode(value).
// Unsupported elements panic according to HashCode's contract.
func Hash(values ...any) uint64 {
	result := uint64(1)
	for _, value := range values {
		result = 31*result + HashCode(value)
	}
	return result
}

func hashComparable(value any) uint64 {
	if IsNil(value) {
		return 0
	}
	v := reflect.ValueOf(value)
	if !v.Comparable() {
		panic("base: HashCode requires a comparable value or Hasher")
	}
	h := fnv.New64a()
	writeValue(h, v)
	return h.Sum64()
}

// Lengths frame variable-sized values. Type tags are unnecessary for the
// equality contract: unequal types are allowed to collide.
func writeValue(h hash.Hash64, v reflect.Value) {
	switch v.Kind() {
	case reflect.Bool:
		var value uint64
		if v.Bool() {
			value = 1
		}
		writeUint64(h, value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		writeUint64(h, uint64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		writeUint64(h, v.Uint())
	case reflect.Float32, reflect.Float64:
		writeUint64(h, floatBits(v.Float()))
	case reflect.Complex64, reflect.Complex128:
		writeUint64(h, floatBits(real(v.Complex())))
		writeUint64(h, floatBits(imag(v.Complex())))
	case reflect.String:
		writeUint64(h, uint64(v.Len()))
		_, _ = h.Write([]byte(v.String()))
	case reflect.Pointer, reflect.Chan, reflect.UnsafePointer:
		writeUint64(h, uint64(v.Pointer()))
	case reflect.Interface:
		if v.IsNil() {
			writeUint64(h, 0)
			return
		}
		writeValue(h, v.Elem())
	case reflect.Array:
		writeUint64(h, uint64(v.Len()))
		for i := 0; i < v.Len(); i++ {
			writeValue(h, v.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).Name != "_" {
				writeValue(h, v.Field(i))
			}
		}
	}
}

func writeUint64(h hash.Hash64, value uint64) {
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[:], value)
	_, _ = h.Write(buffer[:])
}

func floatBits(value float64) uint64 {
	// Go equality equates positive and negative zero.
	if value == 0 {
		return 0
	}
	return math.Float64bits(value)
}
