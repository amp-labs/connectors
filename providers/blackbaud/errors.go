package blackbaud

import (
	"fmt"

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

// ResponseError represents an error response from the blankbaud API.
type ResponseError struct {
	StatusCode int    `json:"statusCode,omitempty"`
	Message    string `json:"message,omitempty"`
	Status     int    `json:"status"`
	Title      string `json:"title"`
	Type       string `json:"type,omitempty"`
	Detail     string `json:"detail,omitempty"`
	Trace_Id   string `json:"trace_id,omitempty"`
	Span_Id    string `json:"span_id,omitempty"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Title)
}
