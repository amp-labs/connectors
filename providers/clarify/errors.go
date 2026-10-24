package clarify

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Errors follow JSON:API: a top-level `errors` array whose entries carry `status` (a string),
// `detail` and, on validation errors, `title` and `source.pointer`.
// https://developer.clarify.ai/docs/api-basics/errors#error-response-format
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"errors"},
		Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
	},
)

type ResponseError struct {
	Errors []struct {
		Status string  `json:"status"`
		Title  string  `json:"title,omitempty"`
		Detail *string `json:"detail"`
		Source *struct {
			Pointer string `json:"pointer"`
		} `json:"source,omitempty"`
	} `json:"errors"`
}

func (r ResponseError) CombineErr(base error) error {
	messages := make([]string, 0, len(r.Errors))

	for _, item := range r.Errors {
		parts := make([]string, 0, 2) // nolint:mnd

		if item.Title != "" {
			parts = append(parts, item.Title)
		}

		if item.Detail != nil && *item.Detail != "" {
			parts = append(parts, *item.Detail)
		}

		if len(parts) != 0 {
			messages = append(messages, strings.Join(parts, ": "))
		}
	}

	if len(messages) == 0 {
		return base
	}

	return fmt.Errorf("%w: %w", base, errors.New(strings.Join(messages, "; "))) // nolint:err113
}
