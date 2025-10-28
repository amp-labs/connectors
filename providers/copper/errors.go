package copper

import (
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"message"},
			Template: func() interpreter.ErrorDescriptor { return &ErrorMessage{} },
		},
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ErrorDetails{} },
		},
	}...,
)

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusUnprocessableEntity: common.ErrBadRequest,
}

type ErrorDetails struct {
	Error string `json:"error"`
}

func (e ErrorDetails) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, e.Error)
}

type ErrorMessage struct {
	Success bool   `json:"success"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e ErrorMessage) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, e.Message)
}
