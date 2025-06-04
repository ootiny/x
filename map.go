package x

func CopyMap[T any](m map[string]T) map[string]T {
	copied := make(map[string]T, len(m))
	for k, v := range m {
		copied[k] = v
	}
	return copied
}

func MergeMap[T any](maps ...map[string]T) map[string]T {
	merged := make(map[string]T)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}
