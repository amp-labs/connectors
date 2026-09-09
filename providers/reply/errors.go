package reply

import (
	"fmt"

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

// ResponseError covers both problem+json shapes Reply returns:
// a plain problem ({"title","status","detail"}) and a validation problem
// carrying an errors array ({"errors":[{"pointer","detail"}],...}).
// https://docs.reply.io/api-reference/schemas/problem-details
type ResponseError struct {
	Title  string        `json:"title"`
	Status int           `json:"status"`
	Detail string        `json:"detail"`
	Errors []ErrorDetail `json:"errors"`
}

type ErrorDetail struct {
	Pointer string `json:"pointer"`
	Detail  string `json:"detail"`
}

func (e ResponseError) CombineErr(base error) error {
	if len(e.Errors) != 0 {
		return fmt.Errorf("%w: %v: %v %v", base, e.Title, e.Errors[0].Pointer, e.Errors[0].Detail)
	}

	if e.Title == "" && e.Detail == "" {
		return base
	}

	return fmt.Errorf("%w: %v: %v", base, e.Title, e.Detail)
}
