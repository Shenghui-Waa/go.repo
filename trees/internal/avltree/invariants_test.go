package avltree

import (
	"cmp"
	"math/rand"
	"testing"
)

func TestBalanceAfterMutations(t *testing.T) {
	tree := NewWith[int, int](cmp.Compare[int])
	rng := rand.New(rand.NewSource(71))
	for range 4000 {
		key := rng.Intn(300)
		if rng.Intn(2) == 0 {
			tree.Put(key, key)
		} else {
			tree.Remove(key)
		}
		count := 0
		var visit func(*Node[int, int], *Node[int, int]) int
		visit = func(n, parent *Node[int, int]) int {
			if n == nil {
				return 0
			}
			count++
			if n.Parent != parent {
				t.Fatal("parent link")
			}
			left, right := visit(n.Children[0], n), visit(n.Children[1], n)
			if left-right > 1 || right-left > 1 {
				t.Fatal("AVL balance")
			}
			return 1 + max(left, right)
		}
		visit(tree.Root, nil)
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
