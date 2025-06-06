package util

func DeduplicateSlice[T comparable](input []T) []T {
	unique := make([]T, 0)
	occurrenceMap := make(map[T]struct{})

	for _, val := range input {
		if _, ok := occurrenceMap[val]; !ok {
			occurrenceMap[val] = struct{}{}
			unique = append(unique, val)
		}
	}

	return unique
}
