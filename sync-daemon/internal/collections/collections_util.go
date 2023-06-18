package collections

func Filter[T any](data []T, f func(T) bool) []T {
	result := make([]T, 0, len(data))
	for _, e := range data {
		if f(e) {
			result = append(result, e)
		}
	}
	return result
}

func Map[T any, V any](data []T, f func(T) V) []V {
	result := make([]V, 0, len(data))
	for _, e := range data {
		result = append(result, f(e))
	}
	return result
}

func CheckUnique[T any, V comparable](data []T, f func(T) V) (bool, *V) {
	s := map[V]bool{}
	for _, item := range data {
		v := f(item)
		if _, exists := s[v]; exists {
			return false, &v
		}
		s[v] = true
	}
	return true, nil
}
