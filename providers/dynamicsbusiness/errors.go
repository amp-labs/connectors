package dynamicsbusiness

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
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Error.Message)
}

var xmlErrorFormats = interpreter.Templates{ // nolint:gochecknoglobals
	func() interpreter.ErrorDescriptor { return &xmlResponseError{} },
}

type xmlResponseError struct {
	Code    string `xml:"code"`
	Message string `xml:"message"`
}

func (r xmlResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Message)
}
