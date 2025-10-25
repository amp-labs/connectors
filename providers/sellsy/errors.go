package sellsy

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

type ResponseError struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details"`
	Context any    `json:"context"`
}

func (e ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, e.Error.Message)
}
