package fireflies

import (
	"errors"
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

// ResponseError represents an error response from the fireflies API.
type ResponseError struct {
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	Friendly   bool   `json:"friendly,omitempty"`
	Locations  any    `json:"locations,omitempty"`
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

// This function uses to check wether the response(200 statuscode) contain error or not.
func checkErrorInResponse(errorArr *ResponseError) error {
	if errorArr == nil {
		return nil
	}

	var errorMsg strings.Builder

	for _, msg := range errorArr.Errors {
		errorMsg.WriteString(msg.Message + "; ")
	}

	return errors.New(strings.TrimSuffix(errorMsg.String(), "; ")) //nolint:err113
}
