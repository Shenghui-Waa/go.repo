package mapcore

import (
	"fmt"
	"math"
	"reflect"

	"github.com/Shenghui-Waa/go.repo/base"
	"github.com/Shenghui-Waa/go.repo/compare"
	rbt "github.com/Shenghui-Waa/go.repo/trees/redblacktree"
)

// ContractError reports unsupported equality/hash operands, not callback bugs.
type ContractError struct{ message string }

func (e ContractError) Error() string { return e.message }

func validate(value any) {
	if value == nil {
		return
	}
	_, equal := value.(base.Equaler)
	_, hash := value.(base.Hasher)
	if equal && hash {
		return
	}
	if !equal && !hash {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Bool, reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
			reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128:
			return
		}
	}
	panic(ContractError{fmt.Sprintf("maps: %T requires Equal and HashCode", value)})
}

func validateHashKey(value any) {
	validate(value)
	if base.IsNil(value) {
		return
	}
	// NaN is not reflexive under native equality and cannot be a usable key.
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		if math.IsNaN(v.Float()) {
			panic(ContractError{"maps: NaN hash key"})
		}
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		if math.IsNaN(real(c)) || math.IsNaN(imag(c)) {
			panic(ContractError{"maps: NaN hash key"})
		}
	}
}

type bucketItem[T, P any] struct {
	key     T
	payload P
}

// index uses collision buckets or the project's typed red-black tree.
type index[T, P any] struct {
	order   compare.Comparator[T]
	tree    *rbt.Tree[T, P]
	buckets map[uint64][]bucketItem[T, P]
}

func newIndex[T, P any](order compare.Comparator[T]) *index[T, P] {
	i := &index[T, P]{order: order}
	if order == nil {
		i.buckets = make(map[uint64][]bucketItem[T, P])
	} else {
		i.tree = rbt.NewWithComparator[T, P](order)
	}
	return i
}
func (i *index[T, P]) validate(key T) {
	if i.order == nil {
		validateHashKey(key)
		return
	}
	if i.order(key, key) != 0 {
		panic(ContractError{"maps: comparator must compare a key equal to itself"})
	}
}
func (i *index[T, P]) get(key T) (P, bool) {
	if i.tree != nil {
		return i.tree.Get(key)
	} else {
		for _, item := range i.buckets[base.HashCode(key)] {
			if base.Equals(item.key, key) {
				return item.payload, true
			}
		}
	}
	var zero P
	return zero, false
}
func (i *index[T, P]) put(key T, value P) {
	if i.tree != nil {
		i.tree.Put(key, value)
		return
	}
	hash := base.HashCode(key)
	bucket := i.buckets[hash]
	for n := range bucket {
		if base.Equals(bucket[n].key, key) {
			bucket[n].payload = value
			return
		}
	}
	i.buckets[hash] = append(bucket, bucketItem[T, P]{key, value})
}
func (i *index[T, P]) remove(key T) {
	if i.tree != nil {
		i.tree.Remove(key)
		return
	}
	hash := base.HashCode(key)
	bucket := i.buckets[hash]
	for n := range bucket {
		if base.Equals(bucket[n].key, key) {
			copy(bucket[n:], bucket[n+1:])
			var zero bucketItem[T, P]
			bucket[len(bucket)-1] = zero
			if len(bucket) == 1 {
				delete(i.buckets, hash)
			} else {
				i.buckets[hash] = bucket[:len(bucket)-1]
			}
			return
		}
	}
}
