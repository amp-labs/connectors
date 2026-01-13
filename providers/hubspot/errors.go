package hubspot

import (
	"errors"
)

var (
	ErrNotArray  = errors.New("results is not an array")
	ErrNotObject = errors.New("result is not an object")
	ErrNotString = errors.New("link is not a string")
)
