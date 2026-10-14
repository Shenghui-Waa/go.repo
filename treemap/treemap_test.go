package treemap

import (
	"math/rand"
	"slices"
	"sort"
	"testing"

	"github.com/Shenghui-Waa/go.repo/compare"
)

func TestTreeMapNavigationAndRange(t *testing.T) {
	mapping := New[int, string](compare.Natural[int])
	for _, key := range []int{5, 1, 9, 3, 7} {
		mapping.Put(key, "value")
	}
	key, value, ok := mapping.LowerEntry(5)
	assertEntry(t, key, value, ok, 3)
	key, value, ok = mapping.FloorEntry(5)
	assertEntry(t, key, value, ok, 5)
	key, value, ok = mapping.CeilingEntry(6)
	assertEntry(t, key, value, ok, 7)
	key, value, ok = mapping.HigherEntry(7)
	assertEntry(t, key, value, ok, 9)
	if got := slices.Collect(mapping.Keys()); !slices.Equal(got, []int{1, 3, 5, 7, 9}) {
		t.Fatalf("keys = %v", got)
	}
	keys := make([]int, 0)
	for key := range mapping.Range(3, true, 9, false) {
		keys = append(keys, key)
	}
	if !slices.Equal(keys, []int{3, 5, 7}) {
		t.Fatalf("range = %v", keys)
	}
}

func TestTreeMapRandomOperationsAndInvariants(t *testing.T) {
	random := rand.New(rand.NewSource(42))
	mapping := New[int, int](compare.Natural[int])
	model := make(map[int]int)
	for step := 0; step < 10000; step++ {
		key := random.Intn(500)
		if random.Intn(3) == 0 {
			got, ok := mapping.Remove(key)
			want, exists := model[key]
			if ok != exists || ok && got != want {
				t.Fatalf("remove %d = %v,%v want %v,%v", key, got, ok, want, exists)
			}
			delete(model, key)
		} else {
			value := random.Int()
			old, replaced := mapping.Put(key, value)
			want, existed := model[key]
			if replaced != existed || replaced && old != want {
				t.Fatalf("put %d mismatch", key)
			}
			model[key] = value
		}
		assertRedBlack(t, mapping)
		if mapping.Len() != len(model) {
			t.Fatalf("len = %d want %d", mapping.Len(), len(model))
		}
	}
	want := make([]int, 0, len(model))
	for key := range model {
		want = append(want, key)
	}
	sort.Ints(want)
	if got := slices.Collect(mapping.Keys()); !slices.Equal(got, want) {
		t.Fatalf("keys mismatch")
	}
}

func assertEntry(t *testing.T, key int, _ string, ok bool, want int) {
	t.Helper()
	if !ok || key != want {
		t.Fatalf("entry = %d,%v want %d,true", key, ok, want)
	}
}

func assertRedBlack(t *testing.T, mapping *TreeMap[int, int]) {
	t.Helper()
	if mapping.root != nil && mapping.root.red {
		t.Fatal("root is red")
	}
	checkBlackHeight(t, mapping.root)
}

func checkBlackHeight(t *testing.T, item *node[int, int]) int {
	t.Helper()
	if item == nil {
		return 1
	}
	if item.red && (isRed(item.left) || isRed(item.right)) {
		t.Fatalf("red node %d has red child", item.key)
	}
	if item.left != nil && item.left.parent != item {
		t.Fatalf("left parent mismatch")
	}
	if item.right != nil && item.right.parent != item {
		t.Fatalf("right parent mismatch")
	}
	left := checkBlackHeight(t, item.left)
	right := checkBlackHeight(t, item.right)
	if left != right {
		t.Fatalf("black height mismatch at %d: %d != %d", item.key, left, right)
	}
	if item.red {
		return left
	}
	return left + 1
}

func FuzzTreeMapMatchesBuiltin(f *testing.F) {
	f.Add([]byte{1, 2, 3, 1, 4})
	f.Fuzz(func(t *testing.T, input []byte) {
		mapping := New[int, int](compare.Natural[int])
		model := make(map[int]int)
		for index, raw := range input {
			key := int(raw % 32)
			if index%3 == 0 {
				mapping.Remove(key)
				delete(model, key)
			} else {
				mapping.Put(key, index)
				model[key] = index
			}
			assertRedBlack(t, mapping)
		}
		if mapping.Len() != len(model) {
			t.Fatalf("len mismatch")
		}
	})
}
