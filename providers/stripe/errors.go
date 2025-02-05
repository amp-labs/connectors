package stripe

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
	http.StatusPaymentRequired: common.ErrBadRequest,
	http.StatusConflict:        common.ErrBadRequest,
}

type ResponseError struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Message       string `json:"message"`
	RequestLogUrl string `json:"request_log_url"` // nolint:tagliatelle,revive
	Type          string `json:"type"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Error.Message)
}
