package clickup

import (
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"ECODE"},
			Template: func() interpreter.ErrorDescriptor { return &CodeError{} },
		},
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &GenericError{} },
		},
	}...,
)

type CodeError struct {
	Err  string `json:"err"`
	Code string `json:"ECODE"`
}

func (e CodeError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v: %v", base, e.Code, e.Err)
}

type GenericError struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	TraceId   int    `json:"trace_id"`
	Timestamp int64  `json:"timestamp"`
}

func (e GenericError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, e.Message)
}
