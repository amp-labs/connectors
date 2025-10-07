package teamwork

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"message"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		},
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

type ResponseMessageError struct {
	Message string `json:"message"`
}

type ResponseError struct {
	Errors []struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
		Meta   any    `json:"meta"`
	} `json:"errors"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Message)
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	list := make([]error, len(r.Errors))

	for index, object := range r.Errors {
		list[index] = fmt.Errorf("%v; %v", object.Title, object.Detail) // nolint:err113
	}

	return fmt.Errorf("%w: %w", base, errors.Join(list...))
}

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusConflict: common.ErrBadRequest,
}
