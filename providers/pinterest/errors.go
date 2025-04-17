package pinterest

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

// ResponseError represents an error response from the Pinterest API.
// Code contains the Pinterest-specific error code and Message contains
// a human-readable description of the error.
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: [%v] %v", base, r.Code, r.Message)
}
