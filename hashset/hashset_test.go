package hashset

import "testing"

func TestHashSetZeroValue(t *testing.T) {
	var set HashSet[int]
	if !set.Add(1) || set.Add(1) {
		t.Fatal("add contract failed")
	}
	if !set.Contains(1) || !set.Remove(1) || set.Remove(1) {
		t.Fatal("set contract failed")
	}
}
