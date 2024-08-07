package intercom

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

/*
Response example:

	{
	  "type": "error.list",
	  "request_id": "00066ltgfertncb684rg",
	  "errors": [
		{
		  "code": "...",
		  "message": "...",
		  "field": "..."
		}
	  ]
	}
*/

var ErrUnknownErrorResponseFormat = errors.New("error response has unexpected format")

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"errors"},
			Template: &ResponseListError{},
		}, {
			MustKeys: nil,
			Template: &ResponseSingleError{},
		},
	}...,
)

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusPaymentRequired:      common.ErrBadRequest,
	http.StatusNotAcceptable:        common.ErrBadRequest,
	http.StatusConflict:             common.ErrBadRequest,
	http.StatusUnsupportedMediaType: common.ErrBadRequest,
	http.StatusUnprocessableEntity:  common.ErrBadRequest,
}

type ResponseListError struct {
	Type      string             `json:"type"`
	RequestId *string            `json:"request_id"` // nolint:tagliatelle
	Errors    []DescriptiveError `json:"errors"`
}

type DescriptiveError struct {
	Code    string  `json:"code"`
	Message *string `json:"message"`
	Field   *string `json:"field"`
}

func (r ResponseListError) CombineErr(base error) error {
	messages := make([]string, len(r.Errors))

	for i, descr := range r.Errors {
		var message string

		if descr.Message != nil {
			message = fmt.Sprintf("[%v]", *descr.Message)
		}

		messages[i] = descr.Code + message
	}

	data := strings.Join(messages, ", ")
	if len(data) == 0 {
		return errors.Join(base, ErrUnknownErrorResponseFormat)
	}

	return fmt.Errorf("%w: %s", base, data)
}

type ResponseSingleError struct {
	Status int64  `json:"status"`
	Error  string `json:"error"`
}

func (r ResponseSingleError) CombineErr(base error) error {
	return fmt.Errorf("%w: %s", base, r.Error)
}
