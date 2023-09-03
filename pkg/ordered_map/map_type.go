package orderedmap

type Map[T any] interface {
	Set(key string, value T)
	Get(key string) (T, bool)
	Delete(key string)
	Keys() []string
	Values() []T
	Len() int
}
