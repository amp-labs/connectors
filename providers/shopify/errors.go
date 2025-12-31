package shopify

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

// errorFormats defines error response patterns for Shopify GraphQL API.
// Shopify returns errors in the standard GraphQL format with an "errors" array.
var errorFormats = interpreter.NewFormatSwitch( //nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

// ResponseError represents an error response from the Shopify GraphQL API.
type ResponseError struct {
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Message    string `json:"message,omitempty"`
	Extensions any    `json:"extensions,omitempty"`
	Locations  any    `json:"locations,omitempty"`
	Path       any    `json:"path,omitempty"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	messages := make([]string, len(r.Errors))
	for i, obj := range r.Errors {
		messages[i] = obj.Message
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ", "))
}
