package capsule

import (
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		},
	}...,
)

type ResponseMessageError struct {
	Message string `json:"message"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Message) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Message)
}
