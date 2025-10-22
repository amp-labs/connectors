package jobber

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Implement error abstraction layers to streamline provider error handling.
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

// ResponseError represents an error response from the Jobber API.
type ResponseError struct {
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Message    string `json:"message,omitempty"`
	Locations  any    `json:"locations,omitempty"`
	Path       any    `json:"path,omitempty"`
	Extensions any    `json:"extensions,omitempty"`
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
