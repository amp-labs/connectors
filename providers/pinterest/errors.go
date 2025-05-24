package pinterest

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
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: [%v] %v", base, r.Code, r.Message)
}
