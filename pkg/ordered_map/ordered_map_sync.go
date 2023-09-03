package orderedmap

import "sync"

type OrderedSyncMap[T any] struct {
	inner Map[T]
	mutex sync.RWMutex
}

func NewSync[T any]() Map[T] {
	return &OrderedSyncMap[T]{
		inner: New[T](),
	}
}

func NewSyncWithCapacity[T any](capacity int) Map[T] {
	return &OrderedSyncMap[T]{
		inner: NewWithCapacity[T](capacity),
	}
}

// Does not copy the map, so the caller must ensure that the map is not modified
func NewSyncFromExisting[T any](m Map[T]) Map[T] {
	return &OrderedSyncMap[T]{
		inner: m,
	}
}

func (m *OrderedSyncMap[T]) Set(key string, value T) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.inner.Set(key, value)
}

func (m *OrderedSyncMap[T]) Get(key string) (T, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.inner.Get(key)
}

func (m *OrderedSyncMap[T]) Delete(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.inner.Delete(key)
}

func (m *OrderedSyncMap[T]) Keys() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.inner.Keys()
}

func (m *OrderedSyncMap[T]) Values() []T {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.inner.Values()
}

func (m *OrderedSyncMap[T]) Len() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.inner.Len()
}
