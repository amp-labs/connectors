package common

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidRawFilter is returned when a RawFilter has a type or filter the connector cannot accept.
	ErrInvalidRawFilter = errors.New("invalid raw filter")

	errFilterNotString = errors.New("must be a string")
	errFilterEmpty     = errors.New("is empty")
)

// RawFilter is a filter in a provider's native syntax.
type RawFilter struct {
	// Type names the filter syntax, e.g. "soql" for Salesforce.
	// Each connector that accepts raw filters documents the types it supports.
	Type string `json:"type"`

	// Filter is the filter expression in that syntax.
	Filter any `json:"filter"`
}

// TypedRawFilter is a connector's own type for one raw filter syntax, such as Salesforce's SOQL filter.
// Its zero value names the syntax it accepts and builds a validated filter from a RawFilter's Filter.
type TypedRawFilter[T any] interface {
	// RawFilterType is the RawFilter.Type this syntax accepts, e.g. "soql".
	RawFilterType() string

	// FromFilter validates a RawFilter's Filter value and returns the typed filter.
	FromFilter(filter any) (T, error)
}

// RawFilterAs resolves a raw filter into the connector's typed filter T, validating it on the way:
// Type must be T's syntax, Filter must be set, and T's FromFilter must accept it.
// Connectors call it wherever they use a raw filter, so validation and conversion happen together.
// Errors wrap ErrInvalidRawFilter.
func RawFilterAs[T TypedRawFilter[T]](rf RawFilter) (T, error) {
	var syntax T

	if rf.Type != syntax.RawFilterType() {
		return syntax, fmt.Errorf("%w: unsupported type %q, expected %q",
			ErrInvalidRawFilter, rf.Type, syntax.RawFilterType())
	}

	// A RawFilter can be set without a Filter.
	if rf.Filter == nil {
		return syntax, fmt.Errorf("%w: %s filter is missing", ErrInvalidRawFilter, rf.Type)
	}

	typed, err := syntax.FromFilter(rf.Filter)
	if err != nil {
		return syntax, fmt.Errorf("%w: %s filter %w", ErrInvalidRawFilter, rf.Type, err)
	}

	return typed, nil
}

// StringFilterValue returns a RawFilter's Filter value as a non-blank string.
// Typed raw filters whose syntax is a plain string use it in FromFilter.
func StringFilterValue(filter any) (string, error) {
	expression, ok := filter.(string)
	if !ok {
		return "", fmt.Errorf("%w, got %T", errFilterNotString, filter)
	}

	if strings.TrimSpace(expression) == "" {
		return "", errFilterEmpty
	}

	return expression, nil
}
