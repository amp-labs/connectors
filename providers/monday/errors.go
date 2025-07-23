package monday

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

type ResponseError struct {
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Extensions struct {
		Code string `json:"code"`
	} `json:"extensions"`
}

// CombineErr joins multiple errors into single golang error.
func (r ResponseError) CombineErr(base error) error {
	messages := make([]string, len(r.Errors))
	for index, detail := range r.Errors {
		messages[index] = detail.Message
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ","))
}
