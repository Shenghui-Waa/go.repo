// Package treemap implements a comparator-ordered red-black tree map.
package treemap

import (
	"iter"

	"github.com/Shenghui-Waa/go.repo/collections"
	"github.com/Shenghui-Waa/go.repo/compare"
	"github.com/Shenghui-Waa/go.repo/iterable"
)

type node[K any, V any] struct {
	key         K
	value       V
	red         bool
	left, right *node[K, V]
	parent      *node[K, V]
}

type TreeMap[K any, V any] struct {
	root    *node[K, V]
	size    int
	version uint64
	less    compare.Comparator[K]
}

func New[K any, V any](less compare.Comparator[K]) *TreeMap[K, V] {
	if less == nil {
		panic("treemap: nil comparator")
	}
	return &TreeMap[K, V]{less: less}
}

func (mapping *TreeMap[K, V]) Comparator() compare.Comparator[K] { return mapping.less }
func (mapping *TreeMap[K, V]) Len() int                          { return mapping.size }
func (mapping *TreeMap[K, V]) IsEmpty() bool                     { return mapping.size == 0 }

func (mapping *TreeMap[K, V]) Get(key K) (V, bool) {
	item := mapping.find(key)
	if item == nil {
		var zero V
		return zero, false
	}
	return item.value, true
}

func (mapping *TreeMap[K, V]) Put(key K, value V) (V, bool) {
	parent := (*node[K, V])(nil)
	current := mapping.root
	order := 0
	for current != nil {
		parent = current
		order = mapping.less(key, current.key)
		if order == 0 {
			old := current.value
			current.value = value
			return old, true
		}
		if order < 0 {
			current = current.left
		} else {
			current = current.right
		}
	}
	item := &node[K, V]{key: key, value: value, red: true, parent: parent}
	if parent == nil {
		mapping.root = item
	} else if order < 0 {
		parent.left = item
	} else {
		parent.right = item
	}
	mapping.fixInsert(item)
	mapping.size++
	mapping.version++
	var zero V
	return zero, false
}

func (mapping *TreeMap[K, V]) fixInsert(item *node[K, V]) {
	for item.parent != nil && item.parent.red {
		parent := item.parent
		grandparent := parent.parent
		if parent == grandparent.left {
			uncle := grandparent.right
			if isRed(uncle) {
				parent.red = false
				uncle.red = false
				grandparent.red = true
				item = grandparent
				continue
			}
			if item == parent.right {
				item = parent
				mapping.rotateLeft(item)
				parent = item.parent
				grandparent = parent.parent
			}
			parent.red = false
			grandparent.red = true
			mapping.rotateRight(grandparent)
		} else {
			uncle := grandparent.left
			if isRed(uncle) {
				parent.red = false
				uncle.red = false
				grandparent.red = true
				item = grandparent
				continue
			}
			if item == parent.left {
				item = parent
				mapping.rotateRight(item)
				parent = item.parent
				grandparent = parent.parent
			}
			parent.red = false
			grandparent.red = true
			mapping.rotateLeft(grandparent)
		}
	}
	mapping.root.red = false
}

func (mapping *TreeMap[K, V]) Remove(key K) (V, bool) {
	removed := mapping.find(key)
	if removed == nil {
		var zero V
		return zero, false
	}
	old := removed.value
	mapping.deleteNode(removed)
	mapping.size--
	mapping.version++
	return old, true
}

func (mapping *TreeMap[K, V]) deleteNode(removed *node[K, V]) {
	replacement := removed
	originalRed := replacement.red
	var child *node[K, V]
	var childParent *node[K, V]

	if removed.left == nil {
		child = removed.right
		childParent = removed.parent
		mapping.transplant(removed, removed.right)
	} else if removed.right == nil {
		child = removed.left
		childParent = removed.parent
		mapping.transplant(removed, removed.left)
	} else {
		replacement = minimum(removed.right)
		originalRed = replacement.red
		child = replacement.right
		if replacement.parent == removed {
			childParent = replacement
			if child != nil {
				child.parent = replacement
			}
		} else {
			childParent = replacement.parent
			mapping.transplant(replacement, replacement.right)
			replacement.right = removed.right
			replacement.right.parent = replacement
		}
		mapping.transplant(removed, replacement)
		replacement.left = removed.left
		replacement.left.parent = replacement
		replacement.red = removed.red
	}
	if !originalRed {
		mapping.fixDelete(child, childParent)
	}
	removed.left = nil
	removed.right = nil
	removed.parent = nil
}

func (mapping *TreeMap[K, V]) fixDelete(item, parent *node[K, V]) {
	for item != mapping.root && !isRed(item) {
		if parent == nil {
			break
		}
		if item == parent.left {
			sibling := parent.right
			if isRed(sibling) {
				sibling.red = false
				parent.red = true
				mapping.rotateLeft(parent)
				sibling = parent.right
			}
			if !isRed(leftOf(sibling)) && !isRed(rightOf(sibling)) {
				setRed(sibling, true)
				item = parent
				parent = item.parent
			} else {
				if !isRed(rightOf(sibling)) {
					setRed(leftOf(sibling), false)
					setRed(sibling, true)
					mapping.rotateRight(sibling)
					sibling = parent.right
				}
				setRed(sibling, parent.red)
				parent.red = false
				setRed(rightOf(sibling), false)
				mapping.rotateLeft(parent)
				item = mapping.root
				parent = nil
			}
		} else {
			sibling := parent.left
			if isRed(sibling) {
				sibling.red = false
				parent.red = true
				mapping.rotateRight(parent)
				sibling = parent.left
			}
			if !isRed(rightOf(sibling)) && !isRed(leftOf(sibling)) {
				setRed(sibling, true)
				item = parent
				parent = item.parent
			} else {
				if !isRed(leftOf(sibling)) {
					setRed(rightOf(sibling), false)
					setRed(sibling, true)
					mapping.rotateLeft(sibling)
					sibling = parent.left
				}
				setRed(sibling, parent.red)
				parent.red = false
				setRed(leftOf(sibling), false)
				mapping.rotateRight(parent)
				item = mapping.root
				parent = nil
			}
		}
	}
	setRed(item, false)
}

func (mapping *TreeMap[K, V]) transplant(old, replacement *node[K, V]) {
	if old.parent == nil {
		mapping.root = replacement
	} else if old == old.parent.left {
		old.parent.left = replacement
	} else {
		old.parent.right = replacement
	}
	if replacement != nil {
		replacement.parent = old.parent
	}
}

func (mapping *TreeMap[K, V]) rotateLeft(pivot *node[K, V]) {
	child := pivot.right
	pivot.right = child.left
	if child.left != nil {
		child.left.parent = pivot
	}
	child.parent = pivot.parent
	if pivot.parent == nil {
		mapping.root = child
	} else if pivot == pivot.parent.left {
		pivot.parent.left = child
	} else {
		pivot.parent.right = child
	}
	child.left = pivot
	pivot.parent = child
}

func (mapping *TreeMap[K, V]) rotateRight(pivot *node[K, V]) {
	child := pivot.left
	pivot.left = child.right
	if child.right != nil {
		child.right.parent = pivot
	}
	child.parent = pivot.parent
	if pivot.parent == nil {
		mapping.root = child
	} else if pivot == pivot.parent.right {
		pivot.parent.right = child
	} else {
		pivot.parent.left = child
	}
	child.right = pivot
	pivot.parent = child
}

func (mapping *TreeMap[K, V]) ContainsKey(key K) bool { return mapping.find(key) != nil }

func (mapping *TreeMap[K, V]) Clear() {
	if mapping.root == nil {
		return
	}
	mapping.root = nil
	mapping.size = 0
	mapping.version++
}

func (mapping *TreeMap[K, V]) All() iter.Seq2[K, V] {
	expectedVersion := mapping.version
	return func(yield func(K, V) bool) {
		stack := make([]*node[K, V], 0)
		current := mapping.root
		for current != nil || len(stack) > 0 {
			for current != nil {
				stack = append(stack, current)
				current = current.left
			}
			current = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if mapping.version != expectedVersion {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(current.key, current.value) {
				return
			}
			current = current.right
		}
	}
}

func (mapping *TreeMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		mapping.All()(func(key K, _ V) bool { return yield(key) })
	}
}

func (mapping *TreeMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		mapping.All()(func(_ K, value V) bool { return yield(value) })
	}
}

func (mapping *TreeMap[K, V]) Backward() iter.Seq2[K, V] {
	expectedVersion := mapping.version
	return func(yield func(K, V) bool) {
		stack := make([]*node[K, V], 0)
		current := mapping.root
		for current != nil || len(stack) > 0 {
			for current != nil {
				stack = append(stack, current)
				current = current.right
			}
			current = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if mapping.version != expectedVersion {
				panic(iterable.ErrConcurrentModification)
			}
			if !yield(current.key, current.value) {
				return
			}
			current = current.left
		}
	}
}

func (mapping *TreeMap[K, V]) FirstKey() (K, bool) {
	item := minimum(mapping.root)
	if item == nil {
		var zero K
		return zero, false
	}
	return item.key, true
}

func (mapping *TreeMap[K, V]) LastKey() (K, bool) {
	item := maximum(mapping.root)
	if item == nil {
		var zero K
		return zero, false
	}
	return item.key, true
}

func (mapping *TreeMap[K, V]) LowerEntry(key K) (K, V, bool) {
	return mapping.entry(mapping.nearby(key, -2))
}
func (mapping *TreeMap[K, V]) FloorEntry(key K) (K, V, bool) {
	return mapping.entry(mapping.nearby(key, -1))
}
func (mapping *TreeMap[K, V]) CeilingEntry(key K) (K, V, bool) {
	return mapping.entry(mapping.nearby(key, 1))
}
func (mapping *TreeMap[K, V]) HigherEntry(key K) (K, V, bool) {
	return mapping.entry(mapping.nearby(key, 2))
}

func (mapping *TreeMap[K, V]) nearby(key K, relation int) *node[K, V] {
	var candidate *node[K, V]
	for current := mapping.root; current != nil; {
		order := mapping.less(current.key, key)
		if order < 0 {
			if relation < 0 {
				candidate = current
			}
			current = current.right
			continue
		}
		if order > 0 {
			if relation > 0 {
				candidate = current
			}
			current = current.left
			continue
		}
		switch relation {
		case -1, 1:
			return current
		case -2:
			return predecessor(current)
		default:
			return successor(current)
		}
	}
	return candidate
}

func predecessor[K any, V any](item *node[K, V]) *node[K, V] {
	if item.left != nil {
		return maximum(item.left)
	}
	for item.parent != nil && item == item.parent.left {
		item = item.parent
	}
	return item.parent
}

func successor[K any, V any](item *node[K, V]) *node[K, V] {
	if item.right != nil {
		return minimum(item.right)
	}
	for item.parent != nil && item == item.parent.right {
		item = item.parent
	}
	return item.parent
}

func (mapping *TreeMap[K, V]) entry(item *node[K, V]) (K, V, bool) {
	if item == nil {
		var key K
		var value V
		return key, value, false
	}
	return item.key, item.value, true
}

func (mapping *TreeMap[K, V]) Range(from K, fromInclusive bool, to K, toInclusive bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if mapping.less(from, to) > 0 {
			return
		}
		mapping.All()(func(key K, value V) bool {
			fromOrder := mapping.less(key, from)
			toOrder := mapping.less(key, to)
			if fromOrder < 0 || (fromOrder == 0 && !fromInclusive) {
				return true
			}
			if toOrder > 0 || (toOrder == 0 && !toInclusive) {
				return false
			}
			return yield(key, value)
		})
	}
}

func (mapping *TreeMap[K, V]) find(key K) *node[K, V] {
	for current := mapping.root; current != nil; {
		order := mapping.less(key, current.key)
		if order == 0 {
			return current
		}
		if order < 0 {
			current = current.left
		} else {
			current = current.right
		}
	}
	return nil
}

func minimum[K any, V any](root *node[K, V]) *node[K, V] {
	for root != nil && root.left != nil {
		root = root.left
	}
	return root
}

func maximum[K any, V any](root *node[K, V]) *node[K, V] {
	for root != nil && root.right != nil {
		root = root.right
	}
	return root
}

func isRed[K any, V any](item *node[K, V]) bool { return item != nil && item.red }
func leftOf[K any, V any](item *node[K, V]) *node[K, V] {
	if item == nil {
		return nil
	}
	return item.left
}
func rightOf[K any, V any](item *node[K, V]) *node[K, V] {
	if item == nil {
		return nil
	}
	return item.right
}
func setRed[K any, V any](item *node[K, V], red bool) {
	if item != nil {
		item.red = red
	}
}

var _ collections.NavigableMap[int, int] = (*TreeMap[int, int])(nil)
