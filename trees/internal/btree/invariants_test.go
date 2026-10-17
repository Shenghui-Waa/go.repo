package btree

import (
	"cmp"
	"math/rand"
	"testing"
)

func TestBalanceAfterMutations(t *testing.T) {
	for _, order := range []int{3, 4, 5, 8, 32} {
		tree := NewWith[int, int](order, cmp.Compare[int])
		rng := rand.New(rand.NewSource(73))
		for range 3000 {
			key := rng.Intn(250)
			if rng.Intn(2) == 0 {
				tree.Put(key, key)
			} else {
				tree.Remove(key)
			}
			count, leafDepth := 0, -1
			var visit func(*Node[int, int], *Node[int, int], int)
			visit = func(n, parent *Node[int, int], depth int) {
				if n == nil {
					return
				}
				if n.Parent != parent {
					t.Fatal("parent link")
				}
				count += len(n.Entries)
				if len(n.Entries) > order-1 {
					t.Fatal("overflow")
				}
				if parent != nil && len(n.Entries) < (order+1)/2-1 {
					t.Fatal("underflow")
				}
				if len(n.Children) == 0 {
					if leafDepth < 0 {
						leafDepth = depth
					}
					if leafDepth != depth {
						t.Fatal("unequal leaf depth")
					}
				} else {
					if len(n.Children) != len(n.Entries)+1 {
						t.Fatal("child count")
					}
					for _, child := range n.Children {
						visit(child, n, depth+1)
					}
				}
			}
			visit(tree.Root, nil, 0)
			if count != tree.Size() {
				t.Fatal("size")
			}
			keys := tree.Keys()
			for i := 1; i < len(keys); i++ {
				if keys[i-1] >= keys[i] {
					t.Fatal("key order")
				}
			}
		}
	}
}
