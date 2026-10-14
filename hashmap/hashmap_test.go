package hashmap

import "testing"

func TestHashMapZeroValueAndOperations(t *testing.T) {
	var mapping HashMap[string, int]
	if _, replaced := mapping.Put("a", 1); replaced {
		t.Fatal("unexpected replacement")
	}
	old, replaced := mapping.Put("a", 2)
	if !replaced || old != 1 {
		t.Fatalf("put = %v, %v", old, replaced)
	}
	value, ok := mapping.Remove("a")
	if !ok || value != 2 || mapping.Len() != 0 {
		t.Fatalf("remove = %v, %v", value, ok)
	}
}
