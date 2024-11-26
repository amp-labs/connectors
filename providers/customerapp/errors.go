package customerapp

import (
	"fmt"
	"net/http"
	"strings"

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
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Detail string `json:"detail"`
	Status string `json:"status"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	details := make([]string, len(r.Errors))
	for i, obj := range r.Errors {
		details[i] = obj.Detail
	}

	return fmt.Errorf("%w: %v", base, strings.Join(details, ", "))
}
