package datautils

type Pair[L, R any] struct {
	Left  L
	Right R
}

func NewPair[L, R any](left L, right R) *Pair[L, R] {
	return &Pair[L, R]{
		Left:  left,
		Right: right,
	}
}
