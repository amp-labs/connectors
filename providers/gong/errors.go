package gong

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
	RequestId string   `json:"requestId"`
	Errors    []string `json:"errors"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, strings.Join(r.Errors, ", "))
}
