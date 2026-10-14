package collections

// AddAll adds every value and reports whether the collection changed.
func AddAll[T any](collection Collection[T], values ...T) bool {
	changed := false
	for _, value := range values {
		if collection.Add(value) {
			changed = true
		}
	}
	return changed
}

// ContainsAll reports whether the collection contains every value.
func ContainsAll[T any](collection Collection[T], values ...T) bool {
	for _, value := range values {
		if !collection.Contains(value) {
			return false
		}
	}
	return true
}

// RemoveAll removes every matching value and reports whether it changed.
func RemoveAll[T any](collection Collection[T], values ...T) bool {
	changed := false
	for _, value := range values {
		for collection.Remove(value) {
			changed = true
		}
	}
	return changed
}
