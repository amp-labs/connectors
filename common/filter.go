package common

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidRawFilter is returned when a RawFilter has a type or filter the connector cannot accept.
var ErrInvalidRawFilter = errors.New("invalid raw filter")

// RawFilter is a filter in a provider's native syntax.
type RawFilter struct {
	// Type names the filter syntax, e.g. "soql" for Salesforce.
	// Each connector that accepts raw filters documents the types it supports.
	Type string `json:"type"`

	// Filter is the filter expression in that syntax.
	Filter any `json:"filter"`
}

// StringValidator checks a string filter expression beyond what AsString checks,
// for rules specific to one connector's syntax. Errors are wrapped in ErrInvalidRawFilter.
type StringValidator func(expression string) error

// AsString returns the filter expression of a raw filter whose syntax is a plain string.
// It checks that Type is expectedType and Filter is set to a non-blank string,
// then runs the connector's own validators on the expression.
// A RawFilter can be set without a Filter; that is rejected like an empty one.
func (rf RawFilter) AsString(expectedType string, validators ...StringValidator) (string, error) {
	if rf.Type != expectedType {
		return "", fmt.Errorf("%w: unsupported type %q, expected %q", ErrInvalidRawFilter, rf.Type, expectedType)
	}

	if rf.Filter == nil {
		return "", fmt.Errorf("%w: %s filter is missing", ErrInvalidRawFilter, expectedType)
	}

	expression, ok := rf.Filter.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s filter must be a string, got %T", ErrInvalidRawFilter, expectedType, rf.Filter)
	}

	if strings.TrimSpace(expression) == "" {
		return "", fmt.Errorf("%w: %s filter is empty", ErrInvalidRawFilter, expectedType)
	}

	for _, validate := range validators {
		if err := validate(expression); err != nil {
			return "", fmt.Errorf("%w: %w", ErrInvalidRawFilter, err)
		}
	}

	return expression, nil
}

// ValidateString checks a raw filter the same way AsString does, without returning the expression.
func (rf RawFilter) ValidateString(expectedType string, validators ...StringValidator) error {
	_, err := rf.AsString(expectedType, validators...)

	return err
}
