package intercom

import (
	"encoding/json"
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

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	var payload ResponseError
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("interpretJSONError general: %w %w", interpreter.ErrUnmarshal, err)
	}

	return payload.combineErr(statusCodeMapping(res, body))
}

func statusCodeMapping(res *http.Response, body []byte) error {
	switch res.StatusCode {
	case http.StatusPaymentRequired:
		return common.ErrBadRequest
	case http.StatusNotAcceptable:
		return common.ErrBadRequest
	case http.StatusConflict:
		return common.ErrBadRequest
	case http.StatusUnsupportedMediaType:
		return common.ErrBadRequest
	case http.StatusUnprocessableEntity:
		return common.ErrBadRequest
	default:
		return interpreter.DefaultStatusCodeMappingToErr(res, body)
	}
}

type ResponseError struct {
	Type      string             `json:"type"`
	RequestId *string            `json:"request_id"` // nolint:tagliatelle
	Errors    []DescriptiveError `json:"errors"`
}

type DescriptiveError struct {
	Code    string  `json:"code"`
	Message *string `json:"message"`
	Field   *string `json:"field"`
}

func (r ResponseError) combineErr(base error) error {
	messages := make([]string, len(r.Errors))

	for i, descr := range r.Errors {
		var message string

		if descr.Message != nil {
			message = fmt.Sprintf("[%v]", *descr.Message)
		}

		messages[i] = descr.Code + message
	}

	data := strings.Join(messages, ",")
	if len(data) == 0 {
		return errors.Join(base, ErrUnknownErrorResponseFormat)
	}

	return fmt.Errorf("%w: %s", base, data)
}
