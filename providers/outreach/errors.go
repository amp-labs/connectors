package outreach

import (
	"errors"
)

var ErrIdMustInt = errors.New("provided record ID must be convertable to integer")
