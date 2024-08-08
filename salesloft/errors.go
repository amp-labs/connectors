package salesloft

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"errors"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseListError{} },
		}, {
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseSingleError{} },
		},
	}...,
)

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusUnprocessableEntity: common.ErrBadRequest,
}

type ResponseListError struct {
	Errors map[string]any `json:"errors"`
}

func (r ResponseListError) CombineErr(base error) error {
	var message string

	data, err := json.Marshal(r.Errors)
	if err != nil {
		message = "failed parsing error response"
	} else {
		message = string(data)
	}

	return fmt.Errorf("%w: %s", base, message)
}

type ResponseSingleError struct {
	Status int64  `json:"status"`
	Err    string `json:"error"`
}

func (r ResponseSingleError) CombineErr(base error) error {
	return fmt.Errorf("%w: %s", base, r.Err)
}
