package mock

import (
	"errors"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrMissingParam   = errors.New("missing required parameter")
)
