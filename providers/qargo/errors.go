package qargo

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Error bodies come in three shapes, per the OpenAPI spec and the rate limit docs:
// 400 ValidationErrorResponse {"message", "errors": [...]}, 422 HTTPValidationError {"detail": [...]}
// and 429 {"error": "Rate limit exceeded"}.
// https://api-docs.qargo.com/docs/section/rate-limits/enforcement
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"message"},
		Template: func() interpreter.ErrorDescriptor { return &messageError{} },
	},
	interpreter.FormatTemplate{
		MustKeys: []string{"detail"},
		Template: func() interpreter.ErrorDescriptor { return &detailError{} },
	},
	interpreter.FormatTemplate{
		MustKeys: []string{"error"},
		Template: func() interpreter.ErrorDescriptor { return &plainError{} },
	},
)

type messageError struct {
	Message string `json:"message"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (e messageError) CombineErr(base error) error {
	details := make([]string, 0, len(e.Errors))

	for _, detail := range e.Errors {
		if detail.Message != "" {
			details = append(details, detail.Message)
		}
	}

	if len(details) == 0 {
		return fmt.Errorf("%w: %s", base, e.Message)
	}

	return fmt.Errorf("%w: %s (%s)", base, e.Message, strings.Join(details, "; "))
}

type detailError struct {
	Detail []struct {
		Loc []any  `json:"loc"`
		Msg string `json:"msg"`
	} `json:"detail"`
}

func (e detailError) CombineErr(base error) error {
	messages := make([]string, 0, len(e.Detail))

	for _, detail := range e.Detail {
		location := make([]string, 0, len(detail.Loc))
		for _, part := range detail.Loc {
			location = append(location, fmt.Sprint(part))
		}

		messages = append(messages, fmt.Sprintf("%s: %s", strings.Join(location, "."), detail.Msg))
	}

	if len(messages) == 0 {
		return base
	}

	return fmt.Errorf("%w: %s", base, strings.Join(messages, "; "))
}

type plainError struct {
	Error string `json:"error"`
}

func (e plainError) CombineErr(base error) error {
	if e.Error == "" {
		return base
	}

	return fmt.Errorf("%w: %s", base, e.Error)
}
