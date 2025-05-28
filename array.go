package x

import "slices"

func ArrayFilter[T any](slice []T, filter func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if filter(item) {
			result = append(result, item)
		}
	}
	return result
}

func ArrayContains[T comparable](slice []T, item T) bool {
	return slices.Contains(slice, item)
}
