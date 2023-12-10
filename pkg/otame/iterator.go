package otame

import (
	"io"
)

var ErrFinished = io.EOF

type Iterator[T any] interface {
	Next() (T, error)
}
