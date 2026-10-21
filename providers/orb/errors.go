package orb

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Errors follow RFC 7807: status, title, type and detail; a 400 adds validation_errors.
// https://docs.withorb.com/api-reference/error-responses
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"title"},
		Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
	},
)

type ResponseError struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Detail string `json:"detail,omitempty"`
	// Items are untyped in the spec, so they are printed as they come.
	ValidationErrors []any `json:"validation_errors,omitempty"`
}

func (r ResponseError) CombineErr(base error) error {
	parts := make([]string, 0, 2+len(r.ValidationErrors)) // nolint:mnd

	for _, part := range []string{r.Title, r.Detail} {
		if part != "" {
			parts = append(parts, part)
		}
	}

	for _, validationError := range r.ValidationErrors {
		parts = append(parts, fmt.Sprint(validationError))
	}

	if len(parts) == 0 {
		return base
	}

	return fmt.Errorf("%w: %s", base, strings.Join(parts, ": "))
}
