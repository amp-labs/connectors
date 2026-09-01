package core

import (
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
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

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusUnprocessableEntity: common.ErrBadRequest,
}

type ResponseError struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v: lob code %v", base, r.Error.Message, r.Error.StatusCode)
}
