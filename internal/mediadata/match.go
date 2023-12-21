package mediadata

import "slices"

type Match[T any] struct {
	ID    string
	Value *T
	Score float64
	Field string
}

// TODO: Reimplement using a slice
type MatchList[T any] map[string]Match[T]

func (l MatchList[T]) Insert(match Match[T]) {
	if m, ok := l[match.ID]; ok {
		if m.Score < match.Score {
			l[match.ID] = match
		}
	} else {
		l[match.ID] = match
	}
}

func (l MatchList[T]) ToSlice() []Match[T] {
	matches := make([]Match[T], 0, len(l))

	for _, match := range l {
		matches = append(matches, match)
	}

	return matches
}

func (l MatchList[T]) Top(n int) []Match[T] {
	slice := l.ToSlice()

	if len(slice) < n {
		return slice
	}

	slices.SortStableFunc(slice, func(x, y Match[T]) int {
		if x.Score < y.Score {
			return 1
		} else if x.Score > y.Score {
			return -1
		}

		return 0
	})

	return slice[:n]
}
