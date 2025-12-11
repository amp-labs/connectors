package jsonquery

import (
	"errors"
	"fmt"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrNullJSON    = errors.New("value of JSON key is null")
	ErrNotArray    = errors.New("JSON value is not an array")
	ErrNotObject   = errors.New("JSON value is not an object")
	ErrNotString   = errors.New("JSON value is not a string")
	ErrNotBool     = errors.New("JSON value is not a boolean")
	ErrNotNumeric  = errors.New("JSON value is not a numeric")
	ErrNotInteger  = errors.New("JSON value is not an integer")
	ErrUnpacking   = errors.New("failed to unpack ajson node")
)

func handleNullNode(key string, optional bool) error {
	if optional {
		return nil
	}

	return formatProblematicKeyError(key, ErrNullJSON)
}

func formatProblematicKeyError(key string, baseErr error) error {
	return fmt.Errorf("problematic key: %v %w", key, baseErr)
}

func createKeyNotFoundErr(key string) error {
	return errors.Join(ErrKeyNotFound, fmt.Errorf("key: [%v]", key)) // nolint:err113
}
