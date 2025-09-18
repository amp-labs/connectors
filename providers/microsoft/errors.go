package microsoft

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
	Error struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		InnerError any    `json:"innerError"`
	} `json:"error"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Error.Message) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Error.Message)
}
