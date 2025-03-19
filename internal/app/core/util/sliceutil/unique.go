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
