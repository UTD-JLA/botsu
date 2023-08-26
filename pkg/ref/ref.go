package ref

func New[T any](v T) *T {
	return &v
}

func DerefArray[T any](v []*T) []T {
	var result []T
	for _, item := range v {
		result = append(result, *item)
	}

	return result
}
