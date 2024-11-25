package constantcontact

import (
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ErrorDetails{} },
		},
	}...,
)

// nolint:tagliatelle
type ErrorDetails struct {
	Key     string `json:"error_key"`
	Message string `json:"error_message"`
}

func (r ErrorDetails) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Message)
}
