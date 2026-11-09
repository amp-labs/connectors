package clickup

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
	Err  string `json:"err"`
	Code string `json:"ECODE"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v: %v", base, r.Code, r.Err)
}
