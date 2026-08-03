package util

func Filter[T any](slice []T, predicate func(T) bool) []T {
	if slice == nil {
		return nil
	}

	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}

	return result
}

func FilterInPlace[T any](s []T, predicate func(T) bool) []T {
	result := s[:0]
	for _, v := range s {
		if predicate(v) {
			result = append(result, v)
		}
	}

	clear(s[len(result):])
	return result
}

func SliceToSet[T comparable](slice []T) map[T]struct{} {
	if slice == nil {
		return nil
	}

	result := make(map[T]struct{}, len(slice))
	for _, v := range slice {
		result[v] = struct{}{}
	}

	return result
}
