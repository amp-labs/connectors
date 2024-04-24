package gong

import (
	"errors"
)

var (
	ErrMissingClient = errors.New("JSON http client not set")
	ErrNotArray      = errors.New("results data is not an array")
	ErrNotObject     = errors.New("record is not an object")
	ErrNotString     = errors.New("next is not a string")
)

func (c *Connector) HandleError(err error) error {
	return err
}
