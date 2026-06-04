package main

func hasDuplicates[T comparable](slice []T) bool {
	seen := make(map[T]struct{})
	for _, item := range slice {
		if _, exists := seen[item]; exists {
			return true
		}

		seen[item] = struct{}{}
	}

	return false
}
