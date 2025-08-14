package breakcold

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
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    Data   `json:"data"`
}

type Data struct {
	Code       string `json:"code"`
	HttpStatus int    `json:"httpStatus"`
	Path       string `json:"path,omitempty"`
}

func (r ResponseError) CombineErr(base error) error {
	// Error field is the safest to return, though not very useful.
	return fmt.Errorf("%w: %v", base, r.Message)
}
