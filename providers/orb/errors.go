package orb

import (
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Errors follow RFC 7807: status, title, type and detail.
// https://docs.withorb.com/api-reference/error-responses
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"title"},
		Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
	},
)

type ResponseError struct {
	Status           int    `json:"status"`
	Title            string `json:"title"`
	Type             string `json:"type"`
	Detail           string `json:"detail"`
	ValidationErrors []any  `json:"validation_errors,omitempty"`
}

func (r ResponseError) CombineErr(base error) error {
	message := r.Title
	if r.Detail != "" {
		message = fmt.Sprintf("%s %s", r.Title, r.Detail)
	}

	if len(r.ValidationErrors) != 0 {
		message = fmt.Sprintf("%s %v", message, r.ValidationErrors)
	}

	return fmt.Errorf("%w: %s", base, message)
}
