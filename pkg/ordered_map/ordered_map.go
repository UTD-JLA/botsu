package orderedmap

type OrderedMap[T any] struct {
	inner map[string]T
	keys  []string
}

func New[T any]() Map[T] {
	return &OrderedMap[T]{
		inner: make(map[string]T),
		keys:  make([]string, 0),
	}
}

func NewWithCapacity[T any](capacity int) Map[T] {
	return &OrderedMap[T]{
		inner: make(map[string]T, capacity),
		keys:  make([]string, 0, capacity),
	}
}

func (m *OrderedMap[T]) Set(key string, value T) {
	if _, ok := m.inner[key]; !ok {
		m.keys = append(m.keys, key)
	}

	m.inner[key] = value
}

func (m *OrderedMap[T]) Get(key string) (T, bool) {
	value, ok := m.inner[key]

	return value, ok
}

func (m *OrderedMap[T]) Delete(key string) {
	delete(m.inner, key)

	for i, k := range m.keys {
		if k == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

func (m *OrderedMap[T]) Keys() []string {
	return m.keys
}

func (m *OrderedMap[T]) Values() []T {
	values := make([]T, 0)

	for _, key := range m.keys {
		values = append(values, m.inner[key])
	}

	return values
}

func (m *OrderedMap[T]) Len() int {
	return len(m.keys)
}
