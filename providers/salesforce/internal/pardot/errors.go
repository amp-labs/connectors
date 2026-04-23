package pardot

import (
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/interpreter"
)

func errorHandlerFunc(rsp *http.Response, body []byte) error {
	return interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle(rsp, body)
}

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Message)
}
