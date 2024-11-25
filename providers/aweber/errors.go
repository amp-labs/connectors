package aweber

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
	HTTPStatus      int    `json:"httpStatus"`
	Code            int    `json:"code"`
	CodeDescription string `json:"codeDescription"`
	Message         string `json:"message"`
	MoreInfo        string `json:"moreInfo"`
	Context         any    `json:"context"`
	UUID            string `json:"uuid"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Message)
}
