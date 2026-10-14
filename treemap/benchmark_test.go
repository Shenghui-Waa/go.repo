package treemap

import (
	"testing"

	"github.com/Shenghui-Waa/go.repo/compare"
)

func BenchmarkTreeMapPutGet(b *testing.B) {
	for iteration := 0; iteration < b.N; iteration++ {
		mapping := New[int, int](compare.Natural[int])
		for key := 0; key < 1000; key++ {
			mapping.Put(key, key)
		}
		for key := 0; key < 1000; key++ {
			mapping.Get(key)
		}
	}
}
