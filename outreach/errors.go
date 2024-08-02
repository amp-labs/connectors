package outreach

import (
	"errors"
)

var (
	ErrMissingClient = errors.New("JSON http client not set")
	ErrNotArray      = errors.New("results data is not an array")
	ErrNotObject     = errors.New("record is not an object")
	ErrNotString     = errors.New("next is not a string")
	ErrEmptyResponse = errors.New("empty response body")
	ErrIdMustInt     = errors.New("provided id must be convertable to int")
	ErrMustJSON      = errors.New("request data couldn't be decoded to json")
)
