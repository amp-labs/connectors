package wealthbox

import (
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Wealthbox error responses follow a simple envelope:
//
//	{"success": false, "errors": "description of what went wrong"}
//
// https://dev.wealthbox.com/#topics-errors
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

type ResponseError struct {
	Success bool `json:"success"`
	Errors  any  `json:"errors"`
}

func (r ResponseError) CombineErr(base error) error {
	if r.Errors == nil {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Errors)
}
