package sliceutil

func Unique[T comparable](vals []T) []T {
	newVals := make([]T, 0)
	uniqueMap := make(map[T]bool, len(vals))
	for _, v := range vals {
		if !uniqueMap[v] {
			newVals = append(newVals, v)
			uniqueMap[v] = true
		}
	}

	return newVals
}

// UniqueFunc returns a slice of unique elements based on the keyFunc
// where it will make comparable key
func UniqueFunc[T any](vals []T, keyFunc func(T) string) []T {
	newVals := make([]T, 0)
	uniqueMap := make(map[string]bool, len(vals))

	for _, v := range vals {
		key := keyFunc(v)
		if !uniqueMap[key] {
			newVals = append(newVals, v)
			uniqueMap[key] = true
		}
	}

	return newVals
}
