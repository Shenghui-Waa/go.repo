package listcore

// Cursor uses node-aware readers so linked traversal is O(n), not O(n squared).
type Cursor[T any, K Kind] struct {
	list    *Core[T, K]
	version uint64
	index   int
	read    func(int) T
}

func (l *Core[T, K]) Cursor() *Cursor[T, K] {
	return &Cursor[T, K]{list: l, version: l.version, index: -1, read: l.store().reader()}
}
func (c *Cursor[T, K]) check() { c.list.check(c.version) }
func (c *Cursor[T, K]) Next() bool {
	c.check()
	if c.index < c.list.Size() {
		c.index++
	}
	return c.index < c.list.Size()
}
func (c *Cursor[T, K]) Prev() bool {
	c.check()
	if c.index >= 0 {
		c.index--
	}
	return c.index >= 0
}
func (c *Cursor[T, K]) Index() int {
	c.check()
	if c.index < 0 || c.index >= c.list.Size() {
		panic("lists: invalid cursor position")
	}
	return c.index
}
func (c *Cursor[T, K]) Value() T    { return c.read(c.Index()) }
func (c *Cursor[T, K]) Begin()      { c.check(); c.index = -1 }
func (c *Cursor[T, K]) End()        { c.check(); c.index = c.list.Size() }
func (c *Cursor[T, K]) First() bool { c.Begin(); return c.Next() }
func (c *Cursor[T, K]) Last() bool  { c.End(); return c.Prev() }
func (c *Cursor[T, K]) NextTo(f func(int, T) bool) bool {
	if f == nil {
		panic("lists: nil predicate")
	}
	for c.Next() {
		found := f(c.Index(), c.Value())
		c.check()
		if found {
			return true
		}
	}
	return false
}
func (c *Cursor[T, K]) PrevTo(f func(int, T) bool) bool {
	if f == nil {
		panic("lists: nil predicate")
	}
	for c.Prev() {
		found := f(c.Index(), c.Value())
		c.check()
		if found {
			return true
		}
	}
	return false
}
