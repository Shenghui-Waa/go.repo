package containers_test

import (
	"slices"
	"testing"

	"github.com/Shenghui-Waa/go.repo/base"
	"github.com/Shenghui-Waa/go.repo/containers"
	"github.com/Shenghui-Waa/go.repo/lists"
	"github.com/Shenghui-Waa/go.repo/maps"
	"github.com/Shenghui-Waa/go.repo/sets"
	"github.com/Shenghui-Waa/go.repo/sets/hashset"
)

// These consumers deliberately omit String, Clear, JSON, equality and hashing.
type basicList struct{ values []int }

func (l *basicList) Values() []int { return slices.Clone(l.values) }
func (l *basicList) Size() int     { return len(l.values) }
func (l *basicList) Empty() bool   { return l.Size() == 0 }
func (l *basicList) Get(i int) (int, bool) {
	if i < 0 || i >= l.Size() {
		return 0, false
	}
	return l.values[i], true
}
func (l *basicList) Add(values ...int) { l.values = append(l.values, values...) }
func (l *basicList) Insert(i int, values ...int) bool {
	if i < 0 || i > l.Size() {
		return false
	}
	l.values = slices.Insert(l.values, i, values...)
	return true
}
func (l *basicList) Set(i, value int) bool {
	if i < 0 || i >= l.Size() {
		return false
	}
	l.values[i] = value
	return true
}
func (l *basicList) Remove(i int) bool {
	if i < 0 || i >= l.Size() {
		return false
	}
	l.values = slices.Delete(l.values, i, i+1)
	return true
}

type basicMap map[int]string

func (m basicMap) Size() int   { return len(m) }
func (m basicMap) Empty() bool { return len(m) == 0 }
func (m basicMap) Values() []string {
	values := make([]string, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}
	return values
}
func (m basicMap) Keys() []int {
	keys := make([]int, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
func (m basicMap) Get(k int) (string, bool) { value, ok := m[k]; return value, ok }
func (m basicMap) Put(k int, value string)  { m[k] = value }
func (m basicMap) Remove(k int) bool {
	_, ok := m[k]
	delete(m, k)
	return ok
}

type basicSet map[int]bool

func (s basicSet) Size() int   { return len(s) }
func (s basicSet) Empty() bool { return len(s) == 0 }
func (s basicSet) Values() []int {
	values := make([]int, 0, len(s))
	for value := range s {
		values = append(values, value)
	}
	return values
}
func (s basicSet) Add(values ...int) int {
	before := len(s)
	for _, value := range values {
		s[value] = true
	}
	return len(s) - before
}
func (s basicSet) Remove(values ...int) int {
	before := len(s)
	for _, value := range values {
		delete(s, value)
	}
	return before - len(s)
}
func (s basicSet) Contains(values ...int) bool {
	for _, value := range values {
		if !s[value] {
			return false
		}
	}
	return true
}

var _ containers.Container[int] = (*basicList)(nil)
var _ lists.List[int] = (*basicList)(nil)
var _ maps.Map[int, string] = basicMap{}
var _ sets.Set[int] = basicSet{}

func TestMinimalContracts(t *testing.T) {
	list := &basicList{}
	list.Add(3, 1, 2)
	if got := containers.GetSortedValues[int](list); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatal(got)
	}
	if !slices.Equal(list.Values(), []int{3, 1, 2}) {
		t.Fatal("source mutated")
	}
	for _, value := range []any{list, basicMap{}, basicSet{}} {
		if _, ok := value.(containers.Clearer); ok {
			t.Fatal("unexpected Clear capability")
		}
		if _, ok := value.(containers.JSONSerializer); ok {
			t.Fatal("unexpected JSON capability")
		}
		if _, ok := value.(base.Hasher); ok {
			t.Fatal("unexpected hash capability")
		}
	}
	m := basicMap{}
	m.Put(1, "value")
	if value, ok := m.Get(1); !ok || value != "value" {
		t.Fatal("basic map")
	}
	left := hashset.New[int]()
	left.Add(1, 2)
	right := basicSet{2: true, 3: true}
	if result := left.Intersection(right); result.Size() != 1 || !result.Contains(2) {
		t.Fatal("set algebra must accept operands without Seq")
	}
}
