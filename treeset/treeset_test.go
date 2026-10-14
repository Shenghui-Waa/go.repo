package treeset

import (
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/compare"
)

func TestTreeSetNavigation(t *testing.T) {
	set := New(compare.Natural[int], 5, 1, 3, 3, 9)
	if set.Len() != 4 {
		t.Fatalf("len = %d", set.Len())
	}
	if got := slices.Collect(set.All()); !slices.Equal(got, []int{1, 3, 5, 9}) {
		t.Fatalf("all = %v", got)
	}
	if value, ok := set.Floor(4); !ok || value != 3 {
		t.Fatalf("floor = %v,%v", value, ok)
	}
	if value, ok := set.Higher(5); !ok || value != 9 {
		t.Fatalf("higher = %v,%v", value, ok)
	}
}
