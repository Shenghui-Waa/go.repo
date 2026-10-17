package redblacktree

import (
	"cmp"
	"math/rand"
	"testing"
)

func TestBalanceAfterMutations(t *testing.T) {
	tree := NewWith[int, int](cmp.Compare[int])
	rng := rand.New(rand.NewSource(72))
	for range 4000 {
		key := rng.Intn(300)
		if rng.Intn(2) == 0 {
			tree.Put(key, key)
		} else {
			tree.Remove(key)
		}
		if tree.Root != nil && tree.Root.color != black {
			t.Fatal("root must be black")
		}
		count := 0
		var visit func(*Node[int, int], *Node[int, int]) int
		visit = func(n, parent *Node[int, int]) int {
			if n == nil {
				return 1
			}
			count++
			if n.Parent != parent {
				t.Fatal("parent link")
			}
			if n.color == red && (nodeColor(n.Left) == red || nodeColor(n.Right) == red) {
				t.Fatal("adjacent red nodes")
			}
			left, right := visit(n.Left, n), visit(n.Right, n)
			if left != right {
				t.Fatal("black height")
			}
			if n.color == black {
				left++
			}
			return left
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
